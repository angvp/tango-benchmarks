FROM ruby:3.4-slim
ENV RAILS_ENV=production BUNDLE_WITHOUT=development:test RAILS_LOG_TO_STDOUT=1
RUN apt-get update -qq && apt-get install -y -qq --no-install-recommends build-essential libpq-dev libyaml-dev curl && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY Gemfile Gemfile.lock ./
RUN bundle install --jobs 4
COPY . .
EXPOSE 8080
ENTRYPOINT ["/app/docker-entrypoint.sh"]
