FROM ruby
LABEL maintainer "@anshumanbh"

RUN apt-get update && apt-get install -y build-essential libpq-dev postgresql-9.4 \
    && gem update --system && gem install gitrob

RUN gem uninstall --force github_api && gem install github_api -v 0.13

ADD .gitrobrc /root/.gitrobrc

RUN echo "user accepted" > /usr/local/bundle/gems/gitrob-1.1.2/agreement.txt

CMD ["gitrob", "server", "--bind-address=0.0.0.0", "--port=9393"]