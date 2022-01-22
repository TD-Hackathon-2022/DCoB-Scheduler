FROM alpine:latest
WORKDIR /opt
COPY ui /opt/ui
COPY DCoB-Scheduler /opt/
EXPOSE 8080
CMD ["./DCoB-Scheduler"]