FROM alpine:latest

LABEL name="jupyter"

RUN apk update && apk add python3 sudo gcc musl-dev linux-headers python3-dev

RUN sudo adduser jupy -D

USER jupy

RUN python3 -m venv /home/jupy/lab

RUN /home/jupy/lab/bin/pip install ipykernel

RUN /home/jupy/lab/bin/python3 -m ipykernel install --user

RUN /home/jupy/lab/bin/pip install jupyterlab

ADD service.py /home/jupy/lab/service.py

# CMD ["/bin/sh","-c",\
#   "/home/jupy/lab/bin/jupyter nbconvert --to notebook --execute --stdin --stdout"]
EXPOSE 7000
CMD ["python","/home/jupy/lab/service.py"]

# CMD ["tail","-f","/dev/null"]
