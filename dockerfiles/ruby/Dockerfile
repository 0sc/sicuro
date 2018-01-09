FROM ubuntu:16.04

# packages required for building rubies with rvm
RUN apt-get update -qqy && apt-get install -qqy \
        bzip2 \
        gawk \
        g++ \
        gcc \
        make \
        libreadline6-dev \
        libyaml-dev \
        libsqlite3-dev \
        sqlite3 \
        autoconf \
        libgmp-dev \
        libgdbm-dev \
        libncurses5-dev \
        automake \
        libtool \
        bison \
        pkg-config \
        libffi-dev \
        git \
        curl \
        nodejs \
        tzdata \
        libpq-dev \
        libmysqlclient-dev \
        qt5-default \
        libqt5webkit5-dev \
        imagemagick \
        libmagickwand-dev \
		jq \
        ssh \
        xvfb \
        && rm -rf /var/lib/apt/lists \
        && truncate -s 0 /var/log/*log

# manually install phantomjs
RUN curl -sL -o - https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-2.1.1-linux-x86_64.tar.bz2 | tar -xjf - -O phantomjs-2.1.1-linux-x86_64/bin/phantomjs > /usr/bin/phantomjs && chmod +x /usr/bin/phantomjs

# install rvm
RUN curl -sSL https://get.rvm.io | bash -s stable
RUN echo 'source /etc/profile.d/rvm.sh' >> ~/.bashrc

# make bundler a default gem
RUN echo bundler >> /usr/local/rvm/gemsets/global.gems

# setup some default flags from rvm (auto install, auto gemset create, quiet curl)
RUN echo "rvm_install_on_use_flag=1\nrvm_gemset_create_on_use_flag=1\nrvm_quiet_curl_flag=1" > ~/.rvmrc

# preinstall some ruby versions
ENV PREINSTALLED_RUBIES "2.3.1 2.4.1"
RUN bash -cl 'for version in $PREINSTALLED_RUBIES; do echo "Now installing Ruby $version"; rvm install $version; rvm cleanup all; done'

COPY docker-entrypoint.sh /usr/local/bin/
RUN ln -s usr/local/bin/docker-entrypoint.sh /
ENTRYPOINT ["docker-entrypoint.sh"]