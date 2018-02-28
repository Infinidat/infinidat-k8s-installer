"# infinidat-k8s-installer" 

infinidat-k8s-installer automates installation of infinidat k8s provisioner. it prompts user for configuration options and deploys all resources related to provisioner.

Steps to install provisioner
1.  Download the INFINIDAT Kubernetes provisioner package in the Kubernetes master node and unzip it.
2.	Execute the installer file using command “./infinidat-k8s-installer”. The installer will then prompt the user to enter values which will set be the initial configmap of the provisioner.
If you press Enter without giving input. Installer will consider default value i.e. Written inside brackets ().
```
user@ubuntu16-master:~$ **./infinidat-k8s-installer -operation=install**
Please enter pool name which should be used to provision volumes/filesystems (default):
Pool1
Please enter provision type which should be used to provision volumes/filesystems <thin/thick> (thin):

Please enter nfs mount options which will be used to mount filesystems (hard,rsize=1048576,wsize=1048576):

Please enter nfs export permissions options which will be used to mount filesystems (RW):
RO
Do you want SSD enabled for filesystem (true) :
true
Should root access be allowed for NFS file systems (true): 
false
Please enter management url for infinibox (http://172.20.212.104/)
http://172.20.212.104/
Please enter username for infinibox :
admin
Please enter password for infinibox :
******
[kubectl create -f installation.yaml]
serviceaccount "infinidat-provisioner" created
clusterrole "infinidat-provisioner-runner" created
clusterrolebinding "run-infinidat-provisioner" created
configmap "provisionerconfig" created
secret "mgmt-api-credentials" created
deployment "infinidat-provisioner" created

```
3.	Verify that the provisioner is running as a pod by executing the “kubectl get pods” command as follows. 
Example:
```
user@ubuntu16-master:~/$ kubectl get pods
NAME                                    READY     STATUS    RESTARTS   AGE
infinidat-provisioner-d78c7f4cb-qj79g   1/1       Running   0          1m
```

In above scenario installer pulls provisoner image from dockerhub. in case your expecting to pull image from your private repo.
you can pass following parameters to installer as follow
```
./infinidat-k8s-installer -imagepath=<reponame/imagename:tag> -imagesecret= secretename
```
secret is optional if your private repo doesnt required any authentication. creation of secret is expained here https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/

4.	You can now create a StorageClass using installer with following parameters. 
Note: installer will help to create yaml for storage you need to create storage class using kubectl create or oc create commands
```
user@ubuntu16-master:~/$ ./infinidat-k8s-installer -operation=storageclass
Which protocol will be supported by this storage class?
 1. NFS
 2. ISCSI
 3. FC
2
Enter Storage Class name :
iscsi-class
Please enter pool name which should be used to provision volumes/filesystems (default):
default
Enter requierd file system for the volume(ext4):
ext3
Do you want to create a write protected volume<true/false>(false):

Network space name as given in Infinibox:<comma seperated names in case of multiple network spaces>:
iSCSI
Please enter provision type which should be used to provision volumes/filesystems <thin/thick> (thin):

Reclaim policy for storageclass<Delete\Retain>(Delete)::

Storage class yaml  iscsi-storageclass20180222153045.yaml   is created in current directory.
```
