FROM alpine:3.5

# Install ca-certificates and openssl with apk
# Add home folder and config folder so we can run locally
RUN apk update && apk add --no-cache ca-certificates && apk add --update openssl && \
    adduser -S -H app && mkdir -p /home/app/.config/gcloud && chown -R app /home/app/

USER app

COPY output/* /
COPY run.sh /run.sh

EXPOSE 8080

ENV LOGXI=*

CMD ["./run.sh"]
