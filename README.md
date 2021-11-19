# directcsi-drive-recovery-v1.4.5
A tool to recover "Terminating" drives from directcsi:v1.4.5 upgrades

There was a major bug found in directcsi:v1.4.4-rhel7 release, where the specific upgrades from 1.4.3-rhel7 to 1.4.4-rhel7 or 1.4.5-rhel7 will make some of the "InUse" drives to "Terminating"

The following are the root-causes for this

- The partition discovery in direct-csi:v1.4.4-rhel7 and direct-csi:v1.4.5-rhel7 will fail to parse the `uevent` file where it expects "PARTN" key to be present.
- After any upgrades from 1.4.3-rhel7 to 1.4.4-rhel7 or 1.4.5-rhel7, such partitions woulbe in "Unavailable" state with an error stating `strconv: cannot parse “”: invalid part num”`
- As a consequence of this discovery failure, any in-use partition drives will be put to "Terminating" state.

The following are the steps to recover from such states

STEP 1: Uninstall direct-csi 

```
kubectl direct-csi uninstall
```

STEP 2: Get the list of drives that are in terminating state

```
kubectl direct-csi drives ls --status "Terminating" -o yaml > terminating-drives.yaml
```

STEP 3: Download the `directcsi-drive-recovery-v1.4.5` binary from this repository and provide executable permissions.

STEP 4: Take a backup of the drive objects that needs to be created. 

(Provide a timestamp to filter the right terminating drives)

```
➜ ./directcsi-drive-recovery-v1.4.5 create terminating-drives.yaml 2021-11-19T08:29:00Z 

 DRIVE         CAPACITY  ALLOCATED  FILESYSTEM  VOLUMES  NODE                 CREATION_TS                    DELETETION_TS                   
 /dev/nvme2n1  8.0 GiB   842 MiB    xfs         8        directcsi-centos7-2  2021-11-19 18:43:07 +0530 IST  2021-11-19 18:58:32 +0530 IST   
 /dev/nvme2n1  8.0 GiB   442 MiB    xfs         4        directcsi-centos7-3  2021-11-19 18:43:08 +0530 IST  2021-11-19 18:58:33 +0530 IST   
 /dev/nvme2n1  8.0 GiB   442 MiB    xfs         4        directcsi-centos7-4  2021-11-19 18:43:08 +0530 IST  2021-11-19 18:58:33 +0530 IST   
 
 ➜ ./directcsi-drive-recovery-v1.4.5 create terminating-drives.yaml 2021-11-19T08:29:00Z --yaml > drives-to-create.yaml
```

STEP 5: Generate deletion manifests for the drives that has to be cleared

```
➜ ./directcsi-drive-recovery-v1.4.5 delete terminating-drives.yaml 2021-11-19T08:29:00Z                                  

 DRIVE         CAPACITY  ALLOCATED  FILESYSTEM  VOLUMES  NODE                 CREATION_TS                    DELETETION_TS                   
 /dev/nvme2n1  8.0 GiB   842 MiB    xfs         8        directcsi-centos7-2  2021-11-19 18:43:07 +0530 IST  2021-11-19 18:58:32 +0530 IST   
 /dev/nvme2n1  8.0 GiB   442 MiB    xfs         4        directcsi-centos7-3  2021-11-19 18:43:08 +0530 IST  2021-11-19 18:58:33 +0530 IST   
 /dev/nvme2n1  8.0 GiB   442 MiB    xfs         4        directcsi-centos7-4  2021-11-19 18:43:08 +0530 IST  2021-11-19 18:58:33 +0530 IST
 
 ➜ ./directcsi-drive-recovery-v1.4.5 delete terminating-drives.yaml 2021-11-19T08:29:00Z --yaml > drives-to-delete.yaml
```

STEP 6: Apply the deletion manifests

```
kubectl apply -f drives-to-delete.yaml
```

STEP 7: Apply the drives that has to be created

```
kubectl create -f drives-to-create.yaml
```

STEP 7: Install `direct-csi:v1.4.5-rhel7-patch` 

```
kubectl direct-csi install --image direct-csi:v1.4.5-rhel7-patch
```

STEP 8: Verify the drive states

```
➜  kubectl direct-csi drives ls      
 DRIVE          CAPACITY  ALLOCATED  FILESYSTEM  VOLUMES  NODE                 ACCESS-TIER  STATUS        
 /dev/nvme1n11  3.6 GiB   42 MiB     xfs         -        directcsi-centos7-1  -            Available     
 /dev/nvme1n12  3.2 GiB   42 MiB     xfs         -        directcsi-centos7-1  -            Available     
 /dev/nvme2n1   8.0 GiB   42 MiB     xfs         -        directcsi-centos7-1  -            Ready         
 /dev/nvme1n11  3.6 GiB   42 MiB     xfs         -        directcsi-centos7-2  -            Available     
 /dev/nvme1n12  3.2 GiB   42 MiB     xfs         -        directcsi-centos7-2  -            Available     
 /dev/nvme2n1   8.0 GiB   842 MiB    xfs         8        directcsi-centos7-2  -            InUse         
 /dev/nvme1n11  3.6 GiB   42 MiB     xfs         -        directcsi-centos7-3  -            Available     
 /dev/nvme1n12  3.2 GiB   42 MiB     xfs         -        directcsi-centos7-3  -            Available     
 /dev/nvme2n1   8.0 GiB   442 MiB    xfs         4        directcsi-centos7-3  -            InUse         
 /dev/nvme1n11  3.6 GiB   42 MiB     xfs         -        directcsi-centos7-4  -            Available     
 /dev/nvme1n12  3.2 GiB   42 MiB     xfs         -        directcsi-centos7-4  -            Available     
 /dev/nvme2n1   8.0 GiB   442 MiB    xfs         4        directcsi-centos7-4  -            InUse   
```
