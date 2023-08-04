
#!/bin/sh
echo "TCPKeepAlive yes" >> /etc/ssh/sshd_config
echo "ClientAliveInterval 60" >> /etc/ssh/sshd_config
echo "ClientAliveCountMax 3" >> /etc/ssh/sshd_config
service sshd restart
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
sudo apt-get install nfs-common -y
sudo rm /lib/systemd/system/nfs-common.service
sudo systemctl daemon-reload
sudo systemctl start nfs-common
sudo systemctl status nfs-common
apt-get update && apt-get install -y \
apt-transport-https ca-certificates curl software-properties-common gnupg2
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
IP=`ifconfig bond0:0 | grep "inet" | awk '{print $2}'`
sudo kubeadm init --apiserver-advertise-address=$IP  --pod-network-cidr=192.168.0.0/16
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
kubectl get nodes
kubectl apply -f https://docs.projectcalico.org/v3.14/manifests/calico.yaml
kubectl get nodes
sudo apt-get install binutils bison gcc make -y
mkdir /export
