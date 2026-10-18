# Creates the tables and the staff account. Runs once, before the workers start.
require "bcrypt"
require_relative "db"

DB.create_table?(:categories) do
  primary_key :id
  String :name, null: false
end
DB.create_table?(:pets) do
  primary_key :id
  String :name, null: false
  Integer :price_cents, null: false
  String :status, null: false
  foreign_key :category_id, :categories, null: false, index: true
  constraint(:price_not_negative) { price_cents >= 0 }
  constraint(:known_status, status: %w[available pending sold])
  index :status
end
DB.create_table?(:customers) do
  primary_key :id
  String :name, null: false
  String :email, null: false
end
DB.create_table?(:orders) do
  primary_key :id
  foreign_key :pet_id, :pets, null: false
  foreign_key :customer_id, :customers, null: false
end
DB.create_table?(:staff) do
  String :username, primary_key: true
  String :password_hash, null: false
end

hash = BCrypt::Password.create(ENV.fetch("BENCH_STAFF_PASSWORD", "bench-staff-password")).to_s
DB[:staff].insert_conflict(target: :username, update: { password_hash: hash }).insert(username: "bench-staff", password_hash: hash)
