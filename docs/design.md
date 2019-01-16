# Stash Design Overview

We are going to make a design overhaul of Stash to simplify backup and recovery process and support some most requested features. This doc will discuss what features stash is going to support and how these features may work.



## Goal

Goal of this new design to support following features:

- Schedule Backup and Recover Volume
- Schedule Backup and Recover Database
- Trigger  Backup Instantly
- Recover in Running Workload
- Default Backup
- Auto Recovery

## Backup Volume

User will be able to backup data from the volumes mounted in a workload.

**What user have to do?**

- Create a `Repository` crd.
- Create a `Backup` crd pointing to targeted workload.

Sample `Repository` crd:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: Repository
metadata:
  name: stash-backup-repo
  namespace: demo
spec:
  backend:
    gcs:
      bucket: stash-backup-repo
      prefix: default/deployment/stash-demo
    storageSecretName: gcs-secret
```

Sample `Backup` crd:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: Backup
metadata:
  name: backup-volume-demo
  namespace: demo
spec:
  schedule: '@every 1h'
  # backupAgent indicates AgentTemplate crd to use for backup the target volume.
  # stash will create some default AgentTemplate  while install to backup/recover various resources.
  # user can also crate their own AgentTemplate to customize backup/recovery
  backupAgent:
    name: stashVolumeBackup
  # repository refers to the Repository crd that hold backend information
  repository:
    name: stash-backup-repo
  # targetRef indicate the target workload that we want to backup
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: stash-demo
  # targetDirectories indicates the directories inside the workload we want to backup
  targetDirectories:
  - /source/data
# retentionPolicies specify the policy to follow to clean old backup snapshots
  retentionPolicy:
    keepLast: 5
    prune: true
  # containerAttributes is an optional field that can be use to set some attributes 
  # like resources, securityContext etc. to backup sidecar container
  containerAttributes:
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
```

**How it will work?**

- Stash will watch for `Backup` crd. When it will find a  `Backup` crd, it will inject a sidecar container to the workload.
- Sidecar container will take periodic backup of target directories.

## Recover Volume

User will be able to recover backed up data  either into a separate volume or into the same workload from where the backup was taken.

### Recover into a Volume

**What user have to do?**

- Create a `Recovery` crd pointing `recoverTo` field to the target volume where the recovered data will be stored.

Sample `Recovery` crd to recover into a volume:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: Recovery
metadata:
  name: recovery-volume-demo
  namespace: demo
spec:
  repository:
    name: stash-backup-repo
    namespace: demo
  recoveryAgent: stashVolumeRecovery
  # paths specifies the directories to recover from the backed up data
  paths:
    - /source/data
  recoverTo:
    # indicates the volume where the recovered data will be stored
    volume:
      mountPath: /source/data
      persistentVolumeClaim:
        claimName: stash-recovered
```

**How it will work?**

- When Stash will find a `Recovery` crd created to recover into a volume, it will launch a Job to recover to that volume.
- The recovery Job will recover data store recovered data to specified volume.

### Recover into same Workload

**What user have to do?**

- Create a `Recovery` crd pointing `recoverTo` field to the workload.

Sample `Recovery` crd to recover into same workload:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: Recovery
metadata:
  name: recovery-database-demo
  namespace: demo
spec:
  repository:
    name: stash-backup-repo
    namespace: demo
  recoveryAgent: pgRecovery
  paths:
    - /source/data
  recoverTo:
    # indicates the running workload where we want to recover
    workload:
      apiVersion: apps/v1
      kind: Deployment
      name: stash-demo
```

**How it will work?**

- When Stash will find a `Recovery` crd created to recover into a workload, it will inject a `init-container` to the targeted workload.
- Then, it will restart the workload.
- The `init-container` will recover data inside the workload.

> **Warning:** Recover in same workload require to restart the workload. So, there will be downtime of the workload.

## Backup Database

User will be able to backup database using Stash.

**What user have to do?**

- Create a `Repository` crd for respective backend.
- Create an `AppBinding` crd which holds connection information for the database. If the database is deployed with [KubeDB](https://kubedb.com/docs/0.9.0/welcome/), `AppBinding` crd will be created automatically for each database.
- Create a `Backup` crd pointing to the `AppBinding` crd.

Sample `AppBinding` crd:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: quick-postgres
  namespace: demo
  labels:
    kubedb.com/kind: Postgres
    kubedb.com/name: quick-postgres
spec:
  clientConfig:
    insecureSkipTLSVerify: true
    service:
      name: quick-postgres
      port: 5432
      scheme: "http"
  secret:
    name: quick-postgres-auth
  type: kubedb.com/postgres
```

Sample `Backup` crd for database backup:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: Backup
metadata:
  name: backup-database-demo
  namespace: demo
spec:
  schedule: '@every 1h'
  # backupAgent indicates AgentTemplate crd to use for backup the target database.
  backupAgent:
    name: pgBackup
  # repository refers to the Repository crd that hold backend information
  repository:
    name: stash-backup-repo
  # targetRef indicates the respective AppBinding crd for target database
  targetRef:
    apiVersion: appcatalog.appscode.com/v1alpha1
    kind: AppBinding
    name: quick-postgres
  # retentionPolicies specify the policy to follow to clean old backup snapshots
  retentionPolicy:
    keepLast: 5
    prune: true
  # containerAttributes is an optional field that can be use to set some attributes
  # such as resources, securityContext etc. to backup sidecar container
  containerAttributes:
    # these arguments will be passed to the backup command.
    # you can use it to backup specific database. by default stash will backup all databases
    args:
      - mydb
      - anotherdb
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
  # podAttributes is an optional field that can be used to set some attributes such as nodeSelector, affinity, toleration etc. for the backup job
  podAttributes:
    nodeSelector:
      cloud.google.com/gke-nodepool: pool-highcpu32
```

**How it will work?**

- When Stash will see a `Backup` crd for database backup, it will lunch  a CronJob to take periodic backup of this database.

## Recover Database

User will be able to initialize a database from backed up snapshot.

**What user have to do?**

- Create a `Recovery` crd with `recoverTo` field pointing to respective `AppBinding` crd of the target database.

Sample `Recovery` crd to recover database:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: Recovery
metadata:
  name: recovery-database-demo
  namespace: demo
spec:
  repository:
    name: stash-backup-repo
    namespace: demo
  recoveryAgent: pgRecovery
  recoverTo:
    # indicates the respective AppBinding crd for target database that we want to initialize from backup
    workload:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: quick-postgres
```

**How it will work?:**

- Stash will lunch a Job to recover the backed up database and initialize target with this recovered this.

## Trigger Backup Instantly

User will be able to trigger a scheduled backup instantly. 

**What user have to do?**

- Create a `BackupTrigger` crd pointing to the target `Backup` crd.

Sample `BackupTrigger` crd:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupTrigger
metadata:
  name: backup-volume-demo-trigger
  namespace: demo
spec:
  # targetBackup indicates the backup crd that we want to trigger to take instant backup
  targetBackup:
    name: backup-volume-demo
```

**How it will work?**

- For scheduled volume backup through sidecar container,  Stash will lunch a `go-routine` to take instant backup.
- For scheduled database backup through CronJob, Stash will lunch another job to take instant backup of the database.

## Default Backup

User will also be able to configure a Default backup for the cluster. So, user will no longer need to create  `Repository` and  `Backup` crd for every workload he want to backup. Instead, he will need to add some annotations to the target workload.

**What user have to do?**

- Create a `BackupTemplate` crd which will hold backend information and backup information.
- Add some annotations to the target workload . If the target is a database then add the annotations to respective `AppBinding` crd.

Sample `BackupTemplate` crd for volume backup:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupTemplate
metadata:
  name: volume-backup-template
spec:
  backend:
    gcs:
      bucket: stash-backup-repo
      prefix: ${metadata.namespace}/${metadata.name} # this prefix template will be used to initialize repository in different directory in backend.
    storageSecretName: gcs-secret
  schedule: '@every 1h'
  backupAgent: stashVolumeBackup
  retentionPolicy:
    name: 'keep-last-5'
    keepLast: 5
    prune: true
```

Sample `BackupTemplate` crd for database backup:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupTemplate
metadata:
  name: database-backup-template
spec:
  backend:
    gcs:
      bucket: stash-backup-repo
      prefix: ${metadata.namespace}/${metadata.name} # this prefix template will be used to initialize repository in different directory in backend.
    storageSecretName: gcs-secret
  schedule: '@every 1h'
  backupAgent: pgBackup
  retentionPolicy:
    name: 'keep-last-5'
    keepLast: 5
    prune: true
```

Sample  workload with annotations for default backup:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stash-demo
  namespace: demo
  labels:
    app: stash-demo
  # if stash find bellow annotations, it will take backup of it.
  annotations:
    stash.appscode.com/backup: true
    stash.appscode.com/targetDirectories: "[/source/data]"
    stash.appscode.com/backupTemplate: "volume-backup-template"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: stash-demo
  template:
    metadata:
      labels:
        app: stash-demo
      name: busybox
    spec:
      containers:
      - args:
        - sleep
        - "3600"
        image: busybox
        imagePullPolicy: IfNotPresent
        name: busybox
        volumeMounts:
        - mountPath: /source/data
          name: source-data
      restartPolicy: Always
      volumes:
      - name: source-data
        configMap:
          name: stash-sample-data
```

Sample `AppBinding` crd with annotations for default backup:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: quick-postgres
  namespace: demo
  labels:
    kubedb.com/kind: Postgres
    kubedb.com/name: quick-postgres
    # if stash find bellow annotations, it will take backup of it.
    annotations:
      stash.appscode.com/backup: true
      stash.appscode.com/backupTemplate: "database-backup-template"
spec:
  clientConfig:
    insecureSkipTLSVerify: true
    service:
      name: quick-postgres
      port: 5432
      scheme: "http"
  secret:
    name: quick-postgres-auth
  type: kubedb.com/postgres
```



**How it will work?**

- Stash will watch the workloads and `AppBinding` crds. When Stash will find an workload/AppBinding crd with these annotations, it will create a `Repository` crd and a `Backup` crd using the information from respective `BackupTemplate`.
- Then, Stash will take normal backup as discussed earlier.



## Auto Recovery

User will be also able to configure an automatic recovery for a particular workload. Each time the workload restart, at first it will perform recovery then original workload's container will start.

**What user have to do?**

- Create a `Recovery` crd with `recoveryPolicy` field set to `Always`

Sample `Recovery` crd configured for auto recovery:

```yam
apiVersion: stash.appscode.com/v1alpha2
kind: Recovery
metadata:
  name: recovery-database-demo
  namespace: demo
spec:
  repository:
    name: stash-backup-repo
    namespace: demo
  recoveryAgent: pgRecovery
  paths:
    - /source/data
  recoverTo:
    # indicates the running workload where we want to recover
    workload:
      apiVersion: apps/v1
      kind: Deployment
      name: stash-demo
  # recoveryPolicy specifies weather to recover only once or recover always when workload restart for a particular Recovery crd.
  # It can be either "IfNotRecovered" or "Always"
  # If "Always" is used, whenever the target workload restart, it will recover first.
  recoveryPolicy: Always
```

**How it will work?**

- When Stash will see a `Recovery` crd configured for auto recovery, it will inject an `init-container` to the target.
- The `init-container` will perform recovery on each restart.



## AgentTemplates

Sample `AgentTemplate` for backup volume:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: AgentTemplate
metadata:
  name: stashVolumeBackup
spec:
  containers:
  - image: appscode/stash:0.8.2
    name:  stash-backup
    args:
    - backup
    - --backup-name=$(BACKUP_NAME)
    - --pushgateway-url=$(PUSH_GATEWAY_URL)
    - --enable-status-subresource=$(ENABLE_STATUS_SUB_RESOURCE)
    - --enable-analytics=$(ENABLE_ANALYSTICS)
```

Sample `AgentTemplate` for recover volume:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: AgentTemplate
metadata:
  name: stashVolumeRecovery
spec:
  containers:
  - image: appscode/stash:0.8.2
    name:  stash-recover
    args:
    - recover
    - --recovery-name=$(RECOVERY_NAME)
    - --pushgateway-url=$(PUSH_GATEWAY_URL)
    - --enable-status-subresource=$(ENABLE_STATUS_SUB_RESOURCE)
    - --enable-analytics=$(ENABLE_ANALYSTICS)
```

Sample `AgentTemplate` for backup PostgreSQL database:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: AgentTemplate
metadata:
  name: pgBackup
spec:
  initContainers:
  # stash-init initialize repository if it does not exist in the backend
  - image: appscode/stash:0.8.2
    name: stash-init
    args:
    - init
    - --backup-name=$(BACKUP_NAME)
  # postgres-tools take backup of the database
  - image: kubedb/postgres-tools:0.9.0
    name:  postgres-tools
    command: ["backup"]
    env:
    - name:  PGPASSWORD
      valueFrom:
        secretKeyRef:
          name: <secret name>
          key: "POSTGRES_PASSWORD"
    - name:  DB_USER
      valueFrom:
        secretKeyRef:
          name: <secret name>
          key: "POSTGRES_USER"
    - name:  DB_HOST
      value: <host url>
    - name: DATA_DIR
      value: <data-dir>
  containers:
  # stash-update-status update Repository crd status and push backup metrics to pushgateway
  - image: appscode/stash:0.8.2
    name:  stash-update-status
    args:
    - update-repo-status
    - --backup-name=$(BACKUP_NAME)
    - --pushgateway-url=$(PUSH_GATEWAY_URL)
    - --enable-status-subresource=$(ENABLE_STATUS_SUB_RESOURCE)
    - --metrics-dir=/tmp/metrics.txt
```

Sample `AgentTemplate` to recover PostgreSQL database:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: AgentTemplate
metadata:
  name: pgRecovery
spec:
  initContainers:
  # postgres-tools recover backup and initialize database
  - image: kubedb/postgres-tools:0.9.0
    name:  postgres-tools
    command: ["recover"]
    env:
      - name:  PGPASSWORD
        valueFrom:
          secretKeyRef:
            name: <secret name>
            key: "POSTGRES_PASSWORD"
      - name:  DB_USER
        valueFrom:
          secretKeyRef:
            name: <secret name>
            key: "POSTGRES_USER"
      - name:  DB_HOST
        value: <host url>
      - name: DATA_DIR
        value: <data-dir>
  containers:
    # stash-update-status update Repository crd status and push recovery metrics to pushgateway
    - image: appscode/stash:0.8.2
      name:  stash-update-status
      args:
      - update-recovery-status
      - --recovery-name=$(RECOVERY_NAME)
      - --pushgateway-url=$(PUSH_GATEWAY_URL)
      - --enable-status-subresource=$(ENABLE_STATUS_SUB_RESOURCE)
      - --metrics-dir=/tmp/metrics.txt
```

