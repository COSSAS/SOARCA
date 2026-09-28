FROM ubuntu:latest 

RUN apt update && apt install  openssh-server sudo -y

RUN sudo useradd -m sshtest

RUN ssh-keygen -q -N "" -f /test

RUN mkdir -p /home/sshtest/.ssh/

RUN cat /test.pub > /home/sshtest/.ssh/authorized_keys

RUN echo "PasswordAuthentication no" >> /etc/ssh/sshd_config

RUN echo "PubkeyAuthentication yes" >> /etc/ssh/sshd_config

RUN sudo service ssh start

EXPOSE 22

CMD ["sudo","/usr/sbin/sshd","-D"]
