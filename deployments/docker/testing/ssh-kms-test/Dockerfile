FROM ubuntu:latest

RUN apt update && apt install  openssh-server sudo -y

RUN sudo useradd sshtest


RUN mkdir -p /home/sshtest/
RUN ssh-keygen -q -N "" -f /home/sshtest/test
RUN mkdir -p /home/sshtest/.ssh/
RUN cat /home/sshtest/test.pub > /home/sshtest/.ssh/authorized_keys
RUN echo "PasswordAuthentication no" >> /etc/ssh/sshd_config
RUN echo "PubkeyAuthentication yes" >> /etc/ssh/sshd_config

RUN service ssh start

EXPOSE 22

CMD ["/usr/sbin/sshd","-D"]
