# XT - Cloud SSH and utils tool
A tool that allows super easy connection to Cloud provider instances through human readable names (using tags).
The tool provides multiple other features related to remote instances, it is written in Go runs like lighting!

# Features
* SSH to various cloud providers and searching by human readable names through tags
* Support multiple providers
* Support multiple environments (i.e. dev, prod, staging)
* Run multiple commands on different servers concurrently 
* Get info on servers and print to table
* Run flows of commands and manipulate data in order to create advanced flows
* Scp to/from servers concurrently

# Requirements
Xt is a wrapper around SSH and SCP, it uses these binaries to run commands instead of trying to write this code which is already running and tested for many years.
* `/usr/bin/ssh` installed
* `/usr/bin/scp` installed

# Installing
```
TODO - write install script
```

# Updates
Xt will auto update in case of new version, you will be asked to confirm at the end of the run

# Usage

## Config
Xt relies on config file environments in order to understand which environments it should connect to, thus before we start using
it we must first configure our environments.
Each environment can be configured with multiple cloud providers.
```
profiles: 
  dev: 
    default: true
    providers: 
      aws: 
        creds-profile: dev
        region: us-west-1
        vpc-id: vpc-112233445566
    ssh: 
      domain: "@dev-example.com"
      user: ec2-user
  prod: 
    providers: 
      aws: 
        creds-profile: prod
        region: us-east-1
        vpc-id: vpc-009988776655
    ssh: 
      domain: "@prod-example.com"
      user: ubuntu
```
* `default: true` marks this profile as default profile to connect and requires no `profile` flag to connect
* `creds-profile` referes to `~/.aws/credentials` profile names
* `domain` can be written as `@ssh-bastion@example.com` in order to provide a final connection string of `<user>@<instanceName>@@ssh-bastion@example.com` thus allowsing connection through bastion or other means of tunneling

## Reasonable Defaults
Xt provides default values but these can be changed via config file, for full config see: [full config example]()

# Connect
SSH to remote instance
```
xt connect -p prod web
Using config file: /Users/user1/.xt/config.yaml
   • Found 2 hosts            
? Instances:  [Use arrows to move, type to filter]
> web-server-prod-1111
  web-server-prod-2222
   • Connecting to web-server-prod-1111...
```
Connect will do the following:
1. will query cloud providers for instances by filter, default is `Name`
2. provide users with the matched instances list for selection
3. ssh to selected instance

To search per different tag provide this via `-t` flag

# Exec
Run commands on remote instances
```
xt exec -p prod -a web ls
Using config file: /Users/user1/.xt/config.yaml
Will run the following: 
Command: ls
Hosts/Pods:
web-server-prod-1111
web-server-prod-2222
Are you sure you want to continue Yes
   • runnning command...       command=ls host=web-server-prod-1111
   • runnning command...       command=ls host=web-server-prod-2222
   • command output            command=ls exit-code=exit status 0 host=web-server-prod-1111
test_file.sh
test_file.txt
   • command output            command=ls exit-code=exit status 0 host=web-server-prod-2222
test_file.sh
test_file.txt
```
* to run command on all matches provide `-a` flag
* to run command without approving it first provide `-f` flag
* to run command and request tty (can be useful for `tail` logs for example), this flag cannot be used together with `-a` flag

# Flows
Run multiple commands pre-configured, manipulate data between each command and parse json outputs to table
## Config
```
flows: 
  connect-pod: 
    - 
      run: "kubectl get pods -o json -l app=microservice"
      output:
        type: json
        root: items
        identifier: metadata.name
    - 
      run: "kubectl exec -it __identifier__ bash"
      tty: true
  print-pods-data: 
    - 
      run: "kubectl get pods -o json -l app=microservice"
      output:
        type: json
        root: items
        print: true
        identifier: metadata.name
        keys: 
        - spec.nodeName
        - metadata.labels.someLabel1
        - metadata.labels.someLabel2
        - spec.containers.0.image
      
```
Running flow:
```
>xt flow -p prod kube-bastion connect-pod
Using config file: /Users/user1/.xt/config.yaml
   • Found 2 hosts            
? Instances: kube-bastion-1111
   • runnning command...       command=kubectl get pods -o json -l app=microservice host=kube-bastion-1111
   • Found 3 hosts            
? Instances:  [Use arrows to move, type to filter]
> microservice-649695b9f8-qnjvq
  microservice-694d64dcfc-nzkl5
  microservice-59ff4998d5-5h6c5
  • runnning command...       command=kubectl exec -it microservice-649695b9f8-qnjvq bash host=kube-bastion-1111

>xt flow -p prod kube-bastion print-pods-data
Using config file: /Users/user1/.xt/config.yaml
   • Found 2 hosts            
? Instances: kube-bastion-1111
   • runnning command...       command=kubectl get pods -o json -l app=microservice host=kube-bastion-1111
• printing info...          command=[kubectl get pods -o json -l app=microservice] host=kube-bastion-1111
NAME                            NODENAME        SOMELABEL1    SOMELABEL2    IMAGE                 
microservice-649695b9f8-qnjvq   node-1-1-1-1    label1        label2        user1/microservice:1.0.0
microservice-694d64dcfc-nzkl5   node-2-2-2-2    label1        label2        user1/microservice:1.0.0
microservice-59ff4998d5-5h6c5   node-3-3-3-3    label1        label2        user1/microservice:1.0.0
```
* `run` is the command to run
* `type` allowed values: `text` `json` default is text => will not try to parse output
* `print` if `true` will print output to stdout
* `identifier` if `type` is `json` then `xt` will try to parse the output and find the provided identifier and save it for the next command which will than substitute the `__identifier__` with the selected name.
Additionally when the command returns it will provide a selectable menu from the found matches if more than 1 match was found.
* `root` if `type` is `json` then `xt` will parse the json starting from the `root`, `root` should be an array of json objects
* `keys` if `type` is `json` then `xt` will parse the json and collect the provided keys, if `print` is `true` then it will print it as a table
* `tty` if set then will attempt to request tty from remote host

# Info
Query cloud provider instances by name and print a formated table of the data
```
xt info -p prod web
Using config file: /Users/user1/.xt/config.yaml
NAME                    INSTANCEID              IMAGE                   TYPE            LIFECYCLE       ARN                                                             PRIVATEIPADDRESS       KEY             LAUNCHTIME                      STATE   AVAILABILITYZONE        PRIVATEDNS                                      SUBNET          VPC          
web-server-1111    i-1111111111     ami-11111   c4.large       spot            arn:aws:iam::1111:instance-profile/web-server-1111      172.14.1.2          prod   2020-05-13 11:10:19 +0000 UTC   enabled us-east-1a              ip-172-14-1-2.us-east-1-1.compute.internal     subnet-111111 vpc-111111
web-server-2222    i-2222222222     ami-11111   c4.large       spot            arn:aws:iam::1111:instance-profile/web-server-2222      172.14.1.3          prod   2020-05-14 05:17:38 +0000 UTC   enabled us-east-1b              ip-172-14-1-3.us-east-1-1.compute.internal     subnet-111111 vpc-111111
```

# SCP
Scp can download/upload from multiple servers.
```
>xt -p prod scp -a web . /home/ubuntu/test_file 
Using config file: /Users/user1/.xt/config.yaml
   • runnning command...       command=web-server-1111_test_file|/home/admin/test_file host=web-server-1111
   • runnning command...       command=web-server-2222_test_file|/home/admin/test_file host=web-server-2222
   • command output            command=web-server-1111_test_file|/home/admin/test_file exit-code=exit status 0 host=web-server-1111
   • command output            command=web-server-2222_test_file|/home/admin/test_file exit-code=exit status 0 host=web-server-2222

>ls
web-server-1111_test_file  web-server-2222_test_file
```
* when using `-a` (during download) files will be copied locally with the server_name as prefix in case multiple files are downloaded with same name
* when using `-ua` (upload to all servers) files will be copied with original name