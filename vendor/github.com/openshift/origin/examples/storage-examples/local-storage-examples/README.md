## Example 1: Deploy nginx application with OSEv3 using local Atomic Host Storage
---


**Summary**: 

This example only uses the hosts local storage and will familiarize users with some basic concepts of pod/container configuration files and OSE operations.  The goal for this task is to successfully deploy a simple nginx pod, map the nginx containers html root directory to a directory on the Atomic Host where html files can be served.  This means when the pod/container is destroyed, the html files will still be available on the Atomic host.  Lastly, to view the running pod from within the OSE Console Web Interface.

- Let’s first set up a simple nginx application that will utilize the local host storage for the web server pages.  This example uses the `local-nginx-pod.json`, Here is the [pod configuration file](local-nginx-pod.json).  Below is a brief explanation of some of the attributes and values used in the file.


*local-nginx-pod.json*

        {
            "apiVersion": "v1",
            "id": "local-nginx",
            "kind": "Pod",
            "metadata": {
                "name": "local-nginx"
            },
            "spec": {
                "containers": [
                     {
                         "name": "local-nginx",
                         "image": "fedora/nginx",
                         "volumeMounts": [
                             {
                                 "mountPath": "/usr/share/nginx/html/test",
                                 "name": "localvol"
                             }
                         ]
                     }
                ],
                "volumes": [
                    {
                         "name": "localvol",
                         "hostPath": {
                            "path": "/opt/data"
                         }
                    }
                ]
            }
        }





_Under volumeMounts_

        mountPath: /usr/share/nginx/html/testlocal   This is the local container mount and storage path (so from container, this path will be created)
        name: localvol                               This is the name of our mount volume and it should match any volumes listed below in the volume section
        

_Under volumes_

        name: localvol     matches name: localvol from volumeMounts
        hostPath           hostPath means local storage on the host, so we have a directory /opt/data local on our host (it must exist)
        path: /opt/data    This is the host mounted path, so you can update or change files from here, like the hello.html or other web pages to be served

`

- Use the OpenShift Console (oc) to deploy the pod


        oc create -f local-nginx-pod.json


        [root@ose1 nginx_local]# oc create -f local-nginx-pod.json 
        pods/local-nginx 

- After a few minutes (this may vary), check and make sure the pod is running

        oc get pods

        [root@ose1 nginx_local]# oc get pods 
        NAME          READY     STATUS    RESTARTS   AGE 
        local-nginx   1/1       Running   0          15m 


- You should now also see the pod running in your OSE Console web Interface  (https://<your master host>:8443/console)  (if AllowAll was enabled, should just be able to login with any id and password)

![OSE nginx](./images/example1_ose_local.png)


- From the OSE Console, Note the “IP on node” and “Node” values, which will tell you what ip and node the nginx application is running on.


- Create a sample (helloworld.html) html page to serve out of the Atomic host `/opt/data` directory.  This is the location of our local mount defined by the mountPath from the `local-nginx-pod.json`.  SSH to the node where the nginx application is running and create the helloworld.html file in the /opt/data directory.
          
         ssh root@<Atomic Node>
         echo “Hello World!  This is being served from local Atomic Host – in /opt/data”  >> /opt/data/helloworld.html  


- ssh into the node using the container_id obtained from “docker ps” command and notice if you go to the nginx root /usr/share/nginx/html/test, you will see the helloworld.html file that we created on our atomic host from the mapped /opt/data directory

        docker ps

        [root@ose2 data]# docker ps 
        CONTAINER ID        IMAGE                         COMMAND             CREATED             STATUS              PORTS               NAMES 
        7f27314f5c3e        fedora/nginx                  "/usr/sbin/nginx"   24 minutes ago      Up 24 minutes                           k8s_local-nginx.ce2651ae_local-nginx_default_d7747b33-45ce-11e5-ae70-52540008f001_3b81240d   
        e3b931561c07        openshift3/ose-pod:v3.0.1.0   "/pod"              27 minutes ago      Up 27 minutes                           k8s_POD.892ec37e_local-nginx_default_d7747b33-45ce-11e5-ae70-52540008f001_ff8363f0      


        docker exec -it 7f27314f5c3e bash

        [root@ose2 data]# docker exec -it 7f27314f5c3e bash 
        bash-4.3# cd /usr/share/nginx/html/testlocal/                                       
        bash-4.3# ls 
        hello.html  *helloworld.html* 


- Enter simple curl command from the docker container to serve the page

        curl http://10.1.0.2/testlocal/helloworld.html

        bash-4.3# curl http://10.1.0.2/testlocal/helloworld.html 
        “Hello World! This is being served from local Atomic Host – in /opt/data” 

===

[Next Example - Gluster](../gluster-examples)

===



