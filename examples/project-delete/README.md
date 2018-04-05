# Delete particular user's OpenShift project

This sample shows how you can use OpenShift client-go to delete user projects in bulk.

## What you need to know?

Every OpenShift project that you create has an annotation with key `"openshift.io/requester"` and value being your username. See following snapshot of the OpenShift project created by user with username `developer`.

```yaml
$ oc get project myapp -o yaml
apiVersion: v1
kind: Project
metadata:
  annotations:
    openshift.io/description: ""
    openshift.io/display-name: ""
    openshift.io/requester: developer
...
```

To delete all the projects that a user owns we just need to look for this annotation if it exists and if it does compare the value of the annotation with the username we want to delete projects of.

## Running this example

You need to have access of a running OpenShift cluster. Then put this code file in a separate directory and also setup the vendor dependency. Run as follows:

```console
$ go run main.go -username developer
2018/04/05 23:03:46 deleting project myapp
2018/04/05 23:03:46 deleting project test1
2018/04/05 23:03:46 deleting project test10
2018/04/05 23:03:46 deleting project test2
2018/04/05 23:03:46 deleting project test3
2018/04/05 23:03:46 deleting project test4
2018/04/05 23:03:46 deleting project test5
2018/04/05 23:03:46 deleting project test6
2018/04/05 23:03:46 deleting project test7
2018/04/05 23:03:46 deleting project test8
2018/04/05 23:03:46 deleting project test9
```

Above you can see that all the projects that user `developer` owned are deleted.