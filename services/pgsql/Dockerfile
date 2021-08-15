#
# PostgreSQL
#
# Create with: docker build -t pgsql:13 .
#    docker tag pgsql:13 localhost:5000/pgsql:13
#    docker push localhost:5000/pgsql:13


# Start from simple alpine container
FROM alpine:3.13

#=================================================================================================================
# Install packages
#
#
RUN apk add --no-cache bash=5.1.0-r0 curl=7.77.0-r0 postgresql=13.3-r0 && rm -rf /var/cache/apk/* \
    && mkdir /run/postgresql && chown postgres /run/postgresql
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/"$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"/bin/linux/amd64/kubectl \
    && chmod +x ./kubectl && mv ./kubectl /usr/local/bin
#
#=================================================================================================================

#=================================================================================================================
# Configure PostgreSQL
#
# Run the rest of the commands as the ``postgres`` user created by the ``postgres`` package when it was ``apt-get installed``
USER postgres

COPY --chown=postgres start_pg.sh /run/postgresql/start_pg.sh
COPY --chown=postgres pg_*.sh /run/postgresql/

#
#=================================================================================================================


#=================================================================================================================
# Setup Docker resources
#
# Expose the PostgreSQL port
EXPOSE 5432
# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/var/lib/postgresql"]
# Set the default command to run when starting the container
CMD ["/bin/bash", "/run/postgresql/start_pg.sh"]
#=================================================================================================================
