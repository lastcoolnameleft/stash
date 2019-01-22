# Stash Design Overview

We are going to make a design overhaul of Stash to simplify backup and recovery process and support some most requested features. This doc will discuss what features stash is going to support and how these features may work.



We  have introduced some new crd  such as [StashTemplate](#stashtemplate), [Action](#action) etc. and made whole process more modular to make stash resources inter-operable between different tools.  This might allow to use stash resources as function in serverless concept.

## Goal

Goal of this new design to support following features:

- [Schedule Backup](#backup-workload-data) and [Recover](#recover-workload-data) Workload Data

- [Schedule Backup](#backup-volume) and [Recover](#recover-volume) Volume
- [Schedule Backup](#backup-database) and [Recover](#recover-database) Database
- [Schedule Backup Cluster YAMLs](#cluster-yaml-backup)
- [Trigger  Backup Instantly](#trigger-backup-instantly)
- [Default Backup](#default-backup)
- [Auto Recovery](#auto-recovery)

## Backup Workload Data

User will be able to backup data from  a running workload.

**What user have to do?**

- Create a `Repository` crd.
- Create a `BackupConfiguration` crd pointing to targeted workload.

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

Sample `BackupConfiguration` crd:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupConfiguration
metadata:
  name: backup-workload-data
  namespace: demo
spec:
  schedule: '@every 1h'
  # <no stashTemplate required for sidecar model>
  # repository refers to the Repository crd that hold backend information
  repository:
    name: stash-backup-repo
  # target indicate the target workload that we want to backup
  target:
    workload:
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

- Stash will watch for `BackupCofiguration` crd. When it will find a  `BackupConfiguration` crd, it will inject a `sidecar` container to the workload and start a `cron` for scheduled backup.
- In each schedule, the `cron` will create `BackupInstance` crd.
- The `sidecar` container watches for `BackupInstance` crd. If it find one, it will take backup instantly and update `BackupInstance` status accordingly.

Sample `BackupInstance` crd:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupInstance
metadata:
  name: backup-volume-demo-instance
  namespace: demo
spec:
  # targetBackupConfiguration indicates the BackupConfiguration crd of respective target that we want to backup
  targetBackupConfiguration:
    name: backup-volume-demo
status:
  observedGeneration: 239844#2
  phase: Succeed
  stats:
    snapshot: 40dc1520
    size: 1.720 GiB
    uploaded: 1.200 GiB # upload size can be smaller than original file size if there are some duplicate files
    fileStats:
      new: 5307
      changed: 0
      unmodified: 0
```



## Recover Workload Data

User will be able to recover backed up data  either into a separate volume or into the same workload from where the backup was taken. Here, is an example for recovering into same workload.

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

## Backup Volume

User will be also able to backup stand-alone volumes. This is useful for `ReadOnlyMany` or `ReadWriteMany` type volumes.

**What user have to do?**

- Create a `Repository` crd for respective backend.

- Create a `BackupConfiguration` crd pointing `target` field to the volume.

Sample `BackupConfiguration` crd to backup a PVC:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupConfiguration
metadata:
  name: backup-volume-demo
  namespace: demo
spec:
  schedule: '@every 1h'
  # stashTemplate indicates StashTemplate crd to use for backup the target volume.
  # stash will create some default StashTemplate  while install to backup/recover various resources.
  # user can also crate their own StashTemplate to customize backup/recovery
  stashTemplate: volumeBackup
  # repository refers to the Repository crd that hold backend information
  repository:
    name: stash-backup-repo
  # target indicate the target workload that we want to backup
  target:
    volume:
      mountPath: /source/data
      persistentVolumeClaim:
        claimName: stash-recovered
  # retentionPolicies specify the policy to follow to clean old backup snapshots
  retentionPolicy:
    keepLast: 5
    prune: true
```

**How it will work?**

1. Stash will create a `CronJob` using information of respective `StashTemplate` crd specified by `stashTeplate` field.
2. The `CronJob` will take periodic backup of the target volume.

## Recover Volume

User will be able to recover backed up data  into a  volume.

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
  # stashTemplate indicates StashTemplate crd to use for creating recovery job
  stashTemplate: volumeRecovery
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

- When Stash will find a `Recovery` crd created to recover into a volume, it will launch a Job to recover into that volume.
- The recovery Job will recover and store recovered data to the specified volume.

## Backup Database

User will be able to backup database using Stash.

**What user have to do?**

- Create a `Repository` crd for respective backend.
- Create an `AppBinding` crd which holds connection information for the database. If the database is deployed with [KubeDB](https://kubedb.com/docs/0.9.0/welcome/), `AppBinding` crd will be created automatically for each database.
- Create a `BackupConfiguration` crd pointing to the `AppBinding` crd.

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

Sample `BackupConfiguration` crd for database backup:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupConfiguration
metadata:
  name: backup-database-demo
  namespace: demo
spec:
  schedule: '@every 1h'
  # stashTemplate indicates StashTemplate crd to use for backup the target database.
  stashTemplate: pgBackup
  # repository refers to the Repository crd that hold backend information
  repository:
    name: stash-backup-repo
  # target indicates the respective AppBinding crd for target database
  target:
    workload:
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

- When Stash will see a `BackupConfiguration` crd for database backup, it will lunch  a `CronJob` to take periodic backup of this database.

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
  # stashTemplate indicates StashTemplate crd to use for creating recovery job
  stashTemplate: pgRecovery
  recoverTo:
    # indicates the respective AppBinding crd for target database that we want to initialize from backup
    workload:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: quick-postgres
```

**How it will work?:**

- Stash will lunch a Job to recover the backed up database and initialize target with this recovered data.

## Cluster YAML Backup

User will be able to backup yaml of the cluster resources. However, currently stash will not provide automatic recover cluster from the YAMLs. So, user will have to create them manually.

In future, Stash might be able to backup and recover not only YAMLs but also entire cluster.

**What user have to do?**

- Create a `Repository` crd for respective backend.
- Create a `BackupConfiguration` crd with `stashTemplate` field point to a `StashTemplate` crd that backup cluster.

Sample `BackupConfiguration` crd to backup YAMLs of cluster resources:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupConfiguration
metadata:
  name: cluster-backup-demo
  namespace: demo
spec:
  schedule: '@every 1h'
  # stashTemplate indicates StashTemplate crd to use for backup the cluster.
  stashTemplate: clusterBackup
  # repository refers to the Repository crd that hold backend information
  repository:
    name: stash-backup-repo
  # <no target required for cluster backup>
  # retentionPolicies specify the policy to follow to clean old backup snapshots
  retentionPolicy:
    keepLast: 5
    prune: true
  # podAttributes is an optional field that can be used to set some attributes such as nodeSelector, affinity, toleration etc. for the backup job
  podAttributes:
    # ServiceAccount `stash-cluster-backup` must have read permission of all resources of the cluster
    serviceAccountName: stash-cluster-backup
```

**How it will work?**

- Stash will lunch a `CronJob` using informations of the `StashTemplate` crd specified through `stashTemplate` filed.
- The `CronJob` will take periodic backup of the cluster.

## Trigger Backup Instantly

User will be able to trigger a scheduled backup instantly. 

**What user have to do?**

- Create a `BackupInstance` crd pointing to the target `BackupConfiguration` crd.

Sample `BackupInstance` crd for triggering instant backup:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: BackupInstance
metadata:
  name: backup-volume-demo-instance
  namespace: demo
spec:
  # targetBackupConfiguration indicates the BackupConfiguration crd of respective target that we want to backup
  targetBackupConfiguration:
    name: backup-volume-demo
```

**How it will work?**

- For scheduled  backup through `sidecar` container, the `sidecar` container will take instant backup as it watches for `BackupInstance` crd.
- For scheduled backup through `CronJob`, Stash will lunch another job to take instant backup of the target.

## Default Backup

User will also be able to configure a `default` backup for the cluster. So, user will no longer need to create  `Repository` and  `BackupConfiguration` crd for every workload he want to backup. Instead, he will need to add some annotations to the target workload.

**What user have to do?**

- Create a `DefaultBackupConfiguration` crd which will hold backend information and backup information.
- Add some annotations to the target. If the target is a database then add the annotations to respective `AppBinding` crd.

### Default Backup of Workload Data

Sample `DefaultBackupConfiguration` crd to backup workload data:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: DefaultBackupConfiguration
metadata:
  name: default-workload-data-backup-configuration
spec:
  backend:
    gcs:
      bucket: stash-backup-repo
      prefix: ${metadata.namespace}/${metadata.name} # this prefix template will be used to initialize repository in different directory in backend.
    storageSecretName: gcs-secret
  schedule: '@every 1h'
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
    stash.appscode.com/StashTemplate: "volume-backup-template"
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



### Default Backup of a PVC

Sample `DefaultBackupConfiguration` crd for stand-alone pvc backup:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: DefaultBackupConfiguration
metadata:
  name: default-volume-backup-configuration
spec:
  backend:
    gcs:
      bucket: stash-backup-repo
      prefix: ${metadata.namespace}/${metadata.name} # this prefix template will be used to initialize repository in different directory in backend.
    storageSecretName: gcs-secret
  schedule: '@every 1h'
  stashTemplate: volumeBackup
  retentionPolicy:
    name: 'keep-last-5'
    keepLast: 5
    prune: true
```

Sample PVC with annotation for default backup:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: demo-pvc
  namespace: demo
  # if stash find bellow annotations, it will take backup of it.
  annotations:
    stash.appscode.com/backup: true
    stash.appscode.com/stashTemplate: "volumeBackup"
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
```



### Default Backup of Database

Sample `DefaultBackupConfiguration` crd for database backup:

```yaml
apiVersion: stash.appscode.com/v1alpha2
kind: DefaultBackupConfiguration
metadata:
  name: database-default-backup-configuration
spec:
  backend:
    gcs:
      bucket: stash-backup-repo
      prefix: ${metadata.namespace}/${metadata.name} # this prefix template will be used to initialize repository in different directory in backend.
    storageSecretName: gcs-secret
  schedule: '@every 1h'
  stashTemplate: pgBackup
  retentionPolicy:
    name: 'keep-last-5'
    keepLast: 5
    prune: true
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
      stash.appscode.com/stashTemplate: "pgBackup"
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

- Stash will watch the workloads, volume and `AppBinding` crds. When Stash will find an workload/volume/AppBinding crd with these annotations, it will create a `Repository` crd and a `BackupConfiguration` crd using the information from respective `StashTemplate`.
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



## Action

`Action` are independent single-containered workload specification that perform only single task. For example, [pgBackup](#pgbackup) takes backup a PostgreSQL database and [clusterBackup](#clusterbackup) takes backup of YAMLs of cluster resources. `Action` crd has some variable fields with `$` prefix which hast be resolved while creating respective workload. You can consider these variable fields as input for an `Action` .

Some example `Action` definition is given below:

#### clusterBackup

```yaml
# clusterBackup action backup yamls of all resources of the cluster
apiVersion: stash.appscode.com/v1alpha2
kind: Action
metadata:
  name: clusterBackup
spec:
  container:
    image: appscode/cluster-backup:0.1.0
    name:  cluster-backup
    command: ["backup"]
    args:
    - sanitize=$(sanitize)
    env:
    - name: RESTIC_REPOSITORY
      value: $(repository)
    envFrom:
    - secretRef:
        name: $(storageSecret)
```

#### pgBackup

```yaml
# pgBackup action backup a PostgreSQL database
apiVersion: stash.appscode.com/v1alpha2
kind: Action
metadata:
  name: pgBackup
spec:
  container:
    image: kubedb/postgres-tools:0.9.0
    name:  postgres-tools
    command: ["backup"]
    args: [$(databases)]
    env:
    - name:  PGPASSWORD
      valueFrom:
        secretKeyRef:
          name: $(databaseSecret)
          key: "POSTGRES_PASSWORD"
    - name:  DB_USER
      valueFrom:
        secretKeyRef:
          name: $(databaseSecret)
          key: "POSTGRES_USER"
    - name:  DB_HOST
      value: $(host)
    - name: DATA_DIR
      value: $(dataDir=/tmp/backup)
    - name: RESTIC_REPOSITORY
      value: $(repository)
    envFrom:
    - secretRef:
        name: $(storageSecret)
```

#### pgRecovery

```yaml
# pgBackup action backup a PostgreSQL database
apiVersion: stash.appscode.com/v1alpha2
kind: Action
metadata:
  name: pgRecovery
spec:
  container:
    image: kubedb/postgres-tools:0.9.0
    name:  postgres-tools
    command: ["recover"]
    env:
    - name:  PGPASSWORD
      valueFrom:
        secretKeyRef:
          name: $(databaseSecret)
          key: "POSTGRES_PASSWORD"
    - name:  DB_USER
      valueFrom:
        secretKeyRef:
          name: $(databaseSecret)
          key: "POSTGRES_USER"
    - name:  DB_HOST
      value: $(host)
    - name: DATA_DIR
      value: $(dataDir=/tmp/backup)
    - name: RESTIC_REPOSITORY
      value: $(repository)
    envFrom:
    - secretRef:
        name: $(storageSecret)
```

#### stashInit

```yaml
# stashInit action initialize a repository in the backend and creates a Repository crd
apiVersion: stash.appscode.com/v1alpha2
kind: Action
metadata:
  name: stashInit
spec:
  container:
    image: appscode/stash:0.9.0
    name:  stash-init
    command: ["init"]
    args:
    - repository=$(repoName)
    envFrom:
    - secretRef:
        name: $(storageSecret)
```

## stashUpdateRepo

```yaml
# stashUpdateRepo update Repository and BackupInstance status for respective backup
apiVersion: stash.appscode.com/v1alpha2
kind: Action
metadata:
  name: stashUpdateRepo
spec:
  container:
    image: appscode/stash:0.9.0
    name:  stash-repo-update
    command: ["update-repo"]
    args:
    - repository=$(repoName)
    - backupInstance=$(backupInstantName)
    envFrom:
    - secretRef:
        name: $(storageSecret)
```



## StashTemplate

A complete backup process may need to perform multiple actions. For example, if you want to backup a PostgreSQL database, we need to initialize a `Repository`, then backup the database and finally update `Repository` status to inform backup is completed or push backup metrics to a `pushgateway` . `StashTemplate` specifies these actions sequentially along with their inputs.



We have chosen to break complete backup process into several actions because if a user want to take backup a PostgreSQL database in a Serverless platform, he can just use `pgBackup` action part as a function.

Some sample `StashTemplate` is given below:

#### pgBackup

```yaml
# pgBackup specifies required actions and their inputs to backup a PostgreSQL database
apiVersion: stash.appscode.com/v1alpha2
kind: StashTemplate
metadata:
  name: pgBackup
spec:
  actions:
  - name: stashInit
    inputs:
      repoName: repositoryName # operator will provide these value
      storageSecret: storageSecret
  - name: pgBackup
    inputs:
      databases: databases
      databaseSecret: databaseSecret
      host: hostURL
      repository: resticRepository
      storageSecret: storageSecret
  - name: stashUpdateRepo
    inputs:
      repoName: repoName
      storageSecret: storageSecret
```

#### pgRecovery

```yaml
# pgRecovery specifies required actions and their inputs to recover a PostgreSQL database from backup
apiVersion: stash.appscode.com/v1alpha2
kind: StashTemplate
metadata:
  name: pgRecovery
spec:
  actions:
  - name: pgRecovery
    inputs:
      databases: databases
      databaseSecret: databaseSecret
      host: hostURL
      repository: resticRepository
      storageSecret: storageSecret
  - name: stashUpdateRecovery
    inputs:
      recoveryName: recoveryName
```

#### clusterBackup

```yaml
# clusterBackup specifies required actions and their inputs to backup cluster yaml
apiVersion: stash.appscode.com/v1alpha2
kind: StashTemplate
metadata:
  name: clusterBackup
spec:
  actions:
  - name: stashInit
    inputs:
      repoName: repositoryName # operator will provide these value
      storageSecret: storageSecret
  - name: clusterBackup
    inputs:
      sanitize: <sanitize>
      repository: resticRepository
      storageSecret: storageSecret
  - name: stashUpdateRepo
    inputs:
      repoName: repoName
      storageSecret: storageSecret
```

