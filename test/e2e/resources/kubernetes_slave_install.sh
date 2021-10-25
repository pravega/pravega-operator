#!/bin/sh

sudo apt-get update
echo "update done"
sudo apt-get install nfs-common -y
sudo rm /lib/systemd/system/nfs-common.service
sudo systemctl daemon-reload
sudo systemctl start nfs-common
sudo systemctl status nfs-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) \
    stable"
sudo apt-get update
echo "update done"
sudo apt-get install docker-ce=5:19.03.9~3-0~ubuntu-focal -y
echo "docker install"
sudo systemctl enable --now docker
apt-get update && apt-get install -y \
  apt-transport-https ca-certificates curl software-properties-common gnupg2 make
echo "installed certs"
sudo curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" \
  | sudo tee -a /etc/apt/sources.list.d/kubernetes.list \
  && sudo apt-get update
sudo apt-get update \
  && sudo apt-get install -yq \
  kubelet=1.21.2-00 \
  kubeadm=1.21.2-00 \
  kubernetes-cni
sudo apt-mark hold kubelet kubeadm kubectl
UUID=`cat /etc/fstab | grep swap | awk '{print $1}' | tr -d "#UUID="`
sed -i '2 s/^/#/' /etc/fstab
echo "swapoff UUID=$UUID"
swapoff UUID=$UUID
