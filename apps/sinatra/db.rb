require "sequel"

DB = Sequel.connect(ENV.fetch("DATABASE_URL"), max_connections: 3)
