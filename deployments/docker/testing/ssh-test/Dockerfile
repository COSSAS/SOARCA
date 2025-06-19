FROM ubuntu:latest

RUN apt update && apt install  openssh-server sudo -y

RUN sudo useradd sshtest

RUN echo "sshtest:pdKY77qNxpI5MAizirtjCVOcm0KFKIs" | sudo chpasswd

RUN echo 'PermitRootLogin yes' >> /etc/ssh/sshd_config

RUN service ssh start

EXPOSE 22

CMD ["/usr/sbin/sshd","-D"]