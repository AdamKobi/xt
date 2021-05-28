# XT - Cloud SSH and utils tool
A tool that allows super easy connection to Cloud provider instances through human readable names (using tags).
The tool provides multiple other features related to remote instances, it is written in Go runs like lighting!
Inspired from the amazing [gh](https://github.com/cli/cli), the official GitHub CLI tool. 

# Features
* All commands support searching Cloud providers using human readable tag names
* SSH to remote servers
* Support multiple environments (i.e. dev, prod, staging)
* Run multiple commands on different servers concurrently 
* Get info on servers and print to table
* Run flows of commands and manipulate data in order to create advanced flows
* Upload/Download files concurrently

# Requirements
Xt is a wrapper around SSH and SCP, it uses these binaries to run commands instead of trying to write this code which is already running and tested for many years.
* `/usr/bin/ssh` installed
* `/usr/bin/scp` installed

# Installing
```
curl -s https://raw.githubusercontent.com/AdamKobi/xt/master/scripts/installer.sh | bash -s
```

# Updates
Xt will update when a new version is available, to install use the install command

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
      - name: aws
        creds-profile: dev
        region: us-east-1
        vpc-id: vpc-112233445566
      - name: aws
        creds-profile: dev
        region: eu-west-1
        vpc-id: vpc-987654312876
    ssh: 
      domain: "@dev-example.com"
      user: ec2-user
  prod: 
    providers: 
      - name: aws 
        creds-profile: prod
        region: us-east-1
        vpc-id: vpc-009988776655
    ssh: 
      domain: "@bastion@prod-example.com"
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
❯ xt connect -p prod web
? Hosts:  [Use arrows to move, type to filter]
> web-prod-5f9e
  web-prod-931f
connecting to web-prod-5f9e
```
Connect will do the following:
1. will query cloud providers for instances by filter, default is `Name`
2. provide user with the matched instances list for selection
3. ssh to selected instance

To search per different tag provide this via `-t` flag

# Run
Run commands on remote instances
```
❯ xt run -p dev_old -a web hostname

? Will Execute
$ hostname
On
web-prod-109e
web-prod-5f9e

 Yes
running command on web-prod-109e
running command on web-prod-5f9e
web-prod-5f9e | web-prod-5f9e
web-prod-109e | web-prod-109e
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
    - run: kubectl get pods -ojson -l app=someapp
      root: items
      output_format: json
      selector: name
      keys:
      - name: name
        path: metadata.name
      - name: role
        path: metadata.labels.role
    - run: kubectl exec -it {{.name}} -c someContainer-{{.role}} bash
  print-pods:
    - run: kubectl get pods -ojson -l app=someapp
      root: items
      keys:
      - name: name
        path: metadata.name
      - name: role
        path: metadata.labels.role
      output_format: json
      print: true
      
```
Running flow:
```
>xt flow run -p prod connect-pod bastion
? Instances:  [Use arrows to move, type to filter]
> microservice-649695b9f8-qnjvq
  microservice-694d64dcfc-nzkl5
  microservice-59ff4998d5-5h6c5
microservice-service@microservice-6655bbcbcb-p2l49:/$

>xt flow run -p prod print-pods-data bastion
NAME                            NODENAME        SOMELABEL1    SOMELABEL2                 
microservice-649695b9f8-qnjvq   node-1-1-1-1    label1        label2       
microservice-694d64dcfc-nzkl5   node-2-2-2-2    label1        label2        
microservice-59ff4998d5-5h6c5   node-3-3-3-3    label1        label2
name is: microservice-6655bbcbcb-p2l49 nodeName is: node-1-1-1-1     
```
* `run` is the command to run
* `output_format` allowed values: `text` `json` default is text => will not try to parse output
* `print` if `true` will print output to stdout
* `selector` if `output_format` is `json` then `xt` will try to parse the output and find the provided selector then save it for the next command which will than substitute the `{{.some_selector}}` with the selected name.
Additionally when the command returns it will provide a selectable menu from the found matches if more than 1 match was found.
* `root` if `output_format` is `json` then `xt` will parse the json starting from the `root`, `root` should be an array of json objects
* `keys` if `output_format` is `json` then `xt` will parse the json and collect the provided keys, if `print` is `true` then it will print it as a table. Keys can be used also for subsititution of next commands

Creating a flow
```
❯ xt flow add flow-example
Adding new flow low-example

? Please type command to run remotley ps -ef
? Choose output format text
? Add additional commands? Yes
? Please type command to run remotley kubectl get pods -o json -l app=microservice
? Choose output format  [Use arrows to move, type to filter, ? for more help]
  text
> json
```

Listing flows
```
xt flow list
connect-pod: 
  - run: kubectl get pods -ojson -l app=someapp
    root: items
    output_format: json
    selector: name
    keys:
    - name: name
      path: metadata.name
    - name: role
      path: metadata.labels.role
  - run: kubectl exec -it {{.name}} -c someContainer-{{.role}} bash
print-pods:
  - run: kubectl get pods -ojson -l app=someapp
    root: items
    keys:
    - name: name
      path: metadata.name
    - name: role
      path: metadata.labels.role
    output_format: json
    print: true
```

Deleting flows
```
xt flow delete print-pods
flow print-pods deleted successfully
```
# Info
Query cloud provider instances by name and print a formated table of the data
```
xt info -p prod web
IInstance Name         Instance ID          Type       Image ID               Private IP Address  Public IP Address  Availability Zone  Subnet           Launch Time            Lifecycle
web-prod-109e  i-084516a4ac534109e  m4.xlarge  ami-111111111111111  172.20.235.214      not found          eu-west-1a         subnet-11111111  2021-02-16 20:03:4...  spot
web-prod-5f9e  i-0f0ff242e85235f9e  c5.xlarge  ami-111111111111111  172.20.236.83       not found          eu-west-1b         subnet-11111111  2021-02-09 22:30:4...  spot
```

# File
File can download/upload from multiple servers.
```
>xt file get -a -p dev_old /home/admin/.bashrc . web
running command on web-prod-109e
running command on web-prod-5f9e
.bashrc                                       100% 3773     7.8KB/s   00:00
.bashrc                                       100% 3773     9.9KB/s   00:00

```
* when using `-a` (during download) files will be copied locally with the server_name as prefix in case multiple files are downloaded with same name
* when using `-ua` (upload to all servers) files will be copied with original name