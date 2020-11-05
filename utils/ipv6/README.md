# WARNING
These are draft notes and are not meant to be used blindly.

# Introduction

This document shows how to set up a cluster on AWS that is IPV6 "only" for GPDB 5X and later testing.  It creates a 
GPDB install environment; it is straightforward for you to install the necessary packages to make it a GPDB build 
environment for 5X and later.  You can set up a single,
2-node, 4-node, or 100-node cluster with this.  You can also set-up a dual-stack(IPv4 and IPv6) environment, and of 
course a IPv4 only environment(though we do not explicitly show how to, it should be straightforward based on
this document).  The creation of the IPv6 AWS environment is conceptually independent of the GPDB run environment,
though that GPDB environment set-up is part of the scripts that create the AWS cluster.

We say "only" IPv6 because,  while we do set up the  cluster to not explicitly show IPV4 localhost or routes, we do
not know how to certify the environment as IPV6.  Ideally, we'd run on a hardware setup that matches that of a specific
  customer environment(including router setups).  Furthermore, there appears to be no way to explicitly 
[turn off IPv4 in AWS](https://docs.aws.amazon.com/vpc/latest/userguide/get-started-ipv6.html).  That said, when we
gpinitsystem, the host/address fields in gp_segment_configuration are the ipv6 hostnames, so we think the cluster is
using ipv6 internally.

Note that we spent a long time trying to get [Nimbus](https://confluence.eng.vmware.com/display/DevToolsDocKB/Nimbus+User+Guide)
 to work for IPv6 testing, but could not.  It is a tool used internally by many teams and has cost advantages,
 so perhaps it can be looked into at a later time if we integrate IPv6 testing into our CI. For one-off testing of
 Cluster Management tools on IPv6, AWS worked well.

The cluster will look like this.  As part of the cluster setup, we will turn off IPv4 access between the MDW and SDW machines to create an IPv6
cluster.  But we leave IPv4 access on the JUMPBOX so you can connect to it externally.
```
your local machine   
       <---(IPv4/IPv6) ---> JUMPBOX
                             <---(IPv4 or IPv6) ---> MDW  /\
                             <---(IPv4 or IPv6) ---> SDW1  |
                             <---(IPv4 or IPv6) ---> SDW2  IPv4/IPv6
                                           ...             |
                             <---(IPv4 or IPv6) ---> SDWn  \/


```

Here are some references for containing some details on the below for your reference:

[IPv6 Networking with AWS VPC](https://1billiontech.com/blog_IPv6_Networking_with_AWS_VPC.php)

[Setting up IPv6 on Amazon with Terraform](https://medium.com/@mattias.holmlund/setting-up-ipv6-on-amazon-with-terraform-e14b3bfef577)

## Possible Future Enhancements
- adjust the scripts below to accept a cluster name and number of nodes(instead of hand-editing)
- make the base below its own "AWS image" to speed up the install and creation process.
- adjust the scripts below to name the ipv6 hostnames without `_ipv6`.

## Gotchas
- the default locale for clusters in AWS is UTF-8; the default in CCP is POSIX/C.  So we explicitly set the locale to
C when creating our cluster.
- the code below checks out gpdb 6.11.1 in /home/centos/gpdb on mdw.
- We use user `centos` in the below.
- Once you disable IPv4 on mdw, you won't be able to use the internet.  You can reboot that machine
and run through the steps to remove IPv4 networking from that host.

# Prerequisites

- This document presumes you have financial authorization(or the account you are using does) for what you are doing, so 
talk with your management to make sure you can do so.  A good practice is to stop your cluster if you will not be using
it for a defined period of time.

- You must have access to AWS to get this to work.  To do so, make sure you can 
[login to the AWS management console](https://us-west-1.console.aws.amazon.com/console/home?region=us-west-1#).  You
will likely want to create your cluster in a region close to you.  Keep in mind the AWS console only shows machines
in a given region.  You will also want to generate keys for [AWS Programmatic Access](https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys).
The AWS "Access Key ID" is know as _AWS_ACCESS_KEY_ID_ below, and the private key as _AWS_SECRET_ACCESS_KEY_. Save
those off somewhere on your local machine.

- It is also nice to have Pivnet programmatic access.  This is known as a _pivnet_api_token_ in the below; you can
generate one after you've logged into your [Pivnet/Tanzu Network account](https://network.pivotal.io/).  Under your
User Profile, press _LEGACY API TOKEN_.  

- You should [install Hashicorp Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) on your local 
machine.  I performed the below with _Terraform v0.13.4_ on MacOSX High Sierra.

- You should [install ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)
on you local machine.   I performed the below with _ansible 2.9.13_ on MacOSX High Sierra.

- We use files in this repository for scripts below. (CHANGE ME)  Download this repo; all files of the form REPO/file refer to a file
in this repository.

# Set up the AWS Cluster

## Set up direnv
This simplifies the process as it provides environment variables that can be accessed via the scripts 
- `cp REPO/.envrc.sample REPO/.envrc`
- edit REPO/.envrc and fill in your _AWS_ACCESS_KEY_ID_, _AWS_SECRET_ACCESS_KEY_ and _pivnet_api_token_ generated 
above.

## Use Terraform to set up the AWS infrastructure
- `terraform init`
- `terraform plan -var dwnode_count=2 -var dwcluster_name="gp_cm_ipv6_1" -auto-approve`
  - _dwnode_count_ is the number of segment hosts in your system; 2 gives mdw, sdw1 and sdw2
  - _dwcluster_name_ is the name of each instance in your cluster in AWS(e.g. "gp_cm_ipv6_1").  Each cluster 
  you create should have a unique name. 
- `terraform apply`

Your cluster is now being created in AWS.  Look at the [Running Instances](https://us-west-1.console.aws.amazon.com/ec2/v2/home?region=us-west-1#Instances:instanceState=running)
in AWS.  If don't see anything after 30 seconds or so, make sure you are in the correct AWS region in the AWS Console.

## Bring in infrastructure values into environment
- edit your _.envrc_ file for the correct number of hosts and for your node name.  Instructions are in that file.
- `direnv allow` to allow your _.envrc_ file to be used.
  - `cd`
  - `cd -` to set up your environment.
- edit the file _gen_ansible_hosts.bash_ for the correct number of hosts.  Instructions are in that file. TODO
- make sure you can connect to the jumpbox:
  - `ping -c2 ${JUMPBOX_IPV4}`
  - `ssh -oStrictHostKeyChecking=no -i files/gp_cm.pem centos@${JUMPBOX_IPV4}`
  - to get JUMPBOX_IPV4 manually, you can run: `jq -r '.resources[] | select(.type == "tls_private_key") | .instances[0].attributes.private_key_pem' terraform.tfstate > files/gp_cm.pem
                                                chmod 600 files/gp_cm.pem`

## Use Ansible to set up the hosts
Note this is a combination of setup of each host in the cluster as well as preparation of each host for installing
GPDB.
### Setup Ansible
- `ansible-galaxy init common`
- `ansible-galaxy init jumpbox`
- `ansible-galaxy init dwcommon`
- `ansible-galaxy init dwcontroller`
- `ansible-galaxy init dwdatanode`

### Run Ansible
You can adjust parameters of the GPDB install, including the number of segments per host, by editing
 _gpdb-vars.yml_.
- `ansible-playbook --inventory-file=ansible_hosts ansible-playbook-all.yml -e @gpdb-vars.yml`


## You now have a functional AWS cluster
You now have a dual-stack cluster with GPDB installed.  You can use the cluster now as a functional AWS dual-stack
cluster, and can set-up the remaining AWS environment for GPDB(exchange keys, etc) yourself.  Or follow the instructions
below.

# Changing the machine to IPv6-only
## Disable IPV4
This will disable IPv4 on AWS; note it is not a permanent removal, due to AWS.  If you reboot, you will notice that
the external IPv4 addresses return, as does IPv4 loopback. 
```
ssh -oStrictHostKeyChecking=no -i files/gp_cm.pem centos@${JUMPBOX_IPV4}
[EDIT THE BWLOW FOR YOUR HOSTNAMES]
for i in mdw_ipv4 sdw1_ipv4 smdw_ipv4; do ssh -oStrictHostKeyChecking=no -o ConnectTimeout=2 $i hostname; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo cat /etc/resolv.conf"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo sed -i -e 's|^search|## search|' -e 's|^nameserver|## nameserver|' /etc/resolv.conf"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo cat /etc/resolv.conf"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo ip a"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo ip a | grep \"inet \""; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo dhclient -v -4 -r"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo mv /var/lib/dhclient/dhclient--ens5.lease /home/centos"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo ip addr del \$(/sbin/ip -4 addr show ens5 | grep inet | awk '{print \$2}') dev ens5"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo ip addr del 127.0.0.1/8 dev lo"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo ip a"; done
for i in mdw_ipv4 sdw1_ipv4 smdw_ipv4; do ssh -oStrictHostKeyChecking=no -o ConnectTimeout=2 $i hostname; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i "sudo hostnamectl set-hostname $i"; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i hostname; done
```
## Test IPv6-only
Run these on the JUMPBOX.  [EDIT THE BWLOW FOR YOUR HOSTNAMES]
### IPv4 should not work
- `for i in  mdw_ipv4 sdw1_ipv4 smdw_ipv4 ; do ssh -oStrictHostKeyChecking=no -o ConnectTimeout=2 $i hostname; done`
### IPv6 should work
- `for i in  mdw_ipv6 sdw1_ipv6 smdw_ipv6 ; do ssh -oStrictHostKeyChecking=no $i hostname; done`

# GPDB Specific Cluster Changes
This is to setup a CCP-like cluster on your machine.  These should all be run on the JUMPBOX.
## Choose your GPDB version
[EDIT THE BWLOW FOR YOUR HOSTNAMES]
```
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i postgres --gp-version; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i sudo ls -al /usr/local; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i sudo chown -R centos:centos /usr/local/greenplum-db-6.11.2; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i sudo chown -R centos:centos /usr/local/greenplum-db-5.28.2; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i 'sudo rm /usr/local/greenplum-db; sudo ln -s /usr/local/greenplum-db-5.28.2 /usr/local/greenplum-db'; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i 'sudo rm /usr/local/greenplum-db; sudo ln -s /usr/local/greenplum-db-6.11.2 /usr/local/greenplum-db'; done
```
## Check GPDB Processes and Cleanup datadirs
[EDIT THE BWLOW FOR YOUR HOSTNAMES]
```
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i ps auxww | grep postgres; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i pkill postgres; done
for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i mkdir -p /data/gpdata/{master,primary,mirror}
```

## Hostname setup
The standard CCP cluster has hostnames mdw, smdw, and sdw1, sdw2,...,sdwN. To change them:
- edit the files/gp_all_hosts_ipv6 to remove the `_ipv6`
- edit the /etc/hosts on all cluster machines to remove the _ipv6 entries.
- `for i in mdw_ipv6 sdw1_ipv6 smdw_ipv6; do ssh -oStrictHostKeyChecking=no $i sudo  hostnamectl set-hostname $i| grep postgres; done`

## Set-up keys
- `gpssh-exkeys -f files/gp_all_hosts_ipv6`

## initialize cluster
- add to _segment_host_list_ all the segment hosts(sdw1,...sdwN)
- `gpinitsystem -a --lc-ctype=POSIX -c gpinitsystem_ccp -h segment_host_list`
- look at gp_segment_configuration; you should see all the ipv6 hostnames

# Test
## Install Behave Test environment
- make sure your installation points to 6X as does your GPDB source.
- follow this:
```
   curl https://bootstrap.pypa.io/get-pip.py -s -L -o get-pip.py
   python get-pip.py
   mkdir -p /tmp/py-requirements
   pip --retries 10 install --ignore-installed --prefix /tmp/py-requirements -r $HOME/gpdb/gpMgmt/requirements-dev.txt
   sudo cp -r /tmp/py-requirements/* $GPHOME/ext/python/
```

## 6X Behave tests
Now you Behave should "just work".
 
## 5X Behave tests
We had difficulty getting this to "just work".  One time we built 5X and another time we didn't,
so consider these provisional.

### Install 5X Behave(this might work)
- source your 5X environment but have gpdb source on the 6X branch
- install prerequisites
```
 (source your 5X environment)
 sudo yum install epel-release
 sudo yum install python-pip
 sudo yum install python-devel
 sudo yum groupinstall 'development tools'
 pip --version
```
- `sudo cp -r /tmp/py-requirements/* $GPHOME/ext/python/`
- Make sure you can “ssh ::1” on all machines in the cluster.


### Building GPDB 5X(hopefully not needed)
NOTE: this installs to /usr/local/gpdb
- reboot mdw to re-enable networking
- install pacakges:
  ```
  sudo yum install bison
  sudo yum install zlib-devel
  sudo yum install flex
  sudo yum install readline-devel
  sudo yum install python-devel
- `cd $HOME/gpdb`
- `git checkout 5.28.2`
- `./configure --disable-gpcloud --enable-depend --enable-debug --enable-cassert  --with-python --without-libxml --prefix=/usr/local/gpdb/ --disable-orca CFLAGS=-O0`
- `make`
- `make install -j8 -s`
- `sudo ln -s /usr/local/gpdb /usr/local/greenplum-db`
- go through steps to remove ipv4 networking above
- transfer /usr/local/gpdb to all other nodes and relink /usr/local/greenplum-devel there
- `export PATH=/usr/local/gpdb/ext/python/bin:$PATH`

## Pull down other GPDB versions from Pivnet(if you want to test another version)
NOTE: you might not be able to install two versions of 5X(or 6X) side-by-side but you can 
upgrade to a newer version.
```
ssh -oStrictHostKeyChecking=no -i files/gp_cm.pem centos@${MDW_IPV4}
pivnet login --api-token="<get personalized Tanzu Net API token>"
pivnet download-product-files --product-slug='pivotal-gpdb' --release-version='6.11.2' --product-file-id=798438
```












