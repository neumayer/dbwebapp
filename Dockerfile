FROM scratch
ADD dbwebapp /dbwebapp
EXPOSE 8080
ENTRYPOINT ["/dbwebapp"]
