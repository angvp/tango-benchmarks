require "bcrypt"
require "json"
require "openssl"
require "sinatra/base"
require_relative "db"

class PetShop < Sinatra::Base
  STATUSES = %w[available pending sold].freeze
  MAX_INT = 2_147_483_647
  TOKEN_LIFETIME = 12 * 60 * 60
  SECRET = ENV.fetch("TOKEN_SECRET")

  set :environment, :production
  set :show_exceptions, false
  set :host_authorization, permitted_hosts: []

  before { content_type :json }

  helpers do
    def fail!(status, message)
      halt status, JSON.generate(error: message)
    end

    def json_body
      body = JSON.parse(request.body.read)
      fail!(400, "body is not valid JSON") unless body.is_a?(Hash)
      body
    rescue JSON::ParserError
      fail!(400, "body is not valid JSON")
    end

    def sign(expiry) = OpenSSL::HMAC.hexdigest("SHA256", SECRET, expiry)

    def valid_token?(token)
      expiry, signature = token.to_s.split(".", 2)
      return false unless expiry&.match?(/\A\d+\z/) && signature
      Rack::Utils.secure_compare(signature, sign(expiry)) && Time.now.to_i < expiry.to_i
    end

    def require_token!
      token = request.env["HTTP_AUTHORIZATION"].to_s.delete_prefix("Bearer ")
      fail!(401, "a valid bearer token is required") unless request.env["HTTP_AUTHORIZATION"].to_s.start_with?("Bearer ") && valid_token?(token)
    end

    def id_param(name)
      raw = params[name]
      raw.match?(/\A\d+\z/) && raw.to_i.between?(1, MAX_INT) ? raw.to_i : nil
    end

    def int_param(name)
      raw = params[name]
      return nil if raw.nil? || raw.empty?
      fail!(400, "numeric parameters must be integers") unless raw.match?(/\A-?\d+\z/)
      raw.to_i
    end

    def int?(value) = value.is_a?(Integer)

    # What is wrong with a pet body, or nil.
    def pet_problem(body)
      price = body["price_cents"]
      category = body["category_id"]
      return "name is required" unless body["name"].is_a?(String) && !body["name"].empty?
      return "price_cents must be between 0 and 2147483647" unless int?(price) && price.between?(0, MAX_INT)
      return "status must be available, pending or sold" unless STATUSES.include?(body["status"])
      return "category_id names no category" unless int?(category) && DB[:categories].where(id: category).count == 1
      nil
    end

    def pet_fields(body) = { name: body["name"], price_cents: body["price_cents"], status: body["status"], category_id: body["category_id"] }
  end

  post "/api/login" do
    body = json_body
    staff = DB[:staff].where(username: body["username"].to_s).first
    unless staff && BCrypt::Password.new(staff[:password_hash]) == body["password"].to_s
      fail!(401, "bad credentials")
    end
    expiry = (Time.now.to_i + TOKEN_LIFETIME).to_s
    JSON.generate(token: "#{expiry}.#{sign(expiry)}")
  end

  get "/api/pets" do
    category, min_price, max_price = int_param("category"), int_param("min_price"), int_param("max_price")
    page = int_param("page") || 1
    page_size = int_param("page_size") || 20
    fail!(400, "page must be at least 1 and page_size between 1 and 100") if page < 1 || !page_size.between?(1, 100)
    status = params["status"]
    fail!(400, "unknown status") if status && !status.empty? && !STATUSES.include?(status)

    pets = DB[:pets]
    pets = pets.where(category_id: category) if category
    pets = pets.where(status: status) if status && !status.empty?
    pets = pets.where { price_cents >= min_price } if min_price
    pets = pets.where { price_cents <= max_price } if max_price
    items = pets.order(:id).limit(page_size).offset((page - 1) * page_size).all
    JSON.generate(items: items, page: page, page_size: page_size, total: pets.count)
  end

  get "/api/pets/:id" do
    id = id_param("id")
    pet = id && DB[:pets].where(id: id).first
    fail!(404, "no such pet") unless pet
    JSON.generate(pet)
  end

  post "/api/pets" do
    require_token!
    body = json_body
    problem = pet_problem(body)
    fail!(422, problem) if problem
    pet = DB[:pets].returning.insert(pet_fields(body)).first
    status 201
    JSON.generate(pet)
  end

  put "/api/pets/:id" do
    require_token!
    id = id_param("id")
    fail!(404, "no such pet") unless id && DB[:pets].where(id: id).count == 1
    body = json_body
    problem = pet_problem(body)
    fail!(422, problem) if problem
    JSON.generate(DB[:pets].where(id: id).returning.update(pet_fields(body)).first)
  end

  post "/api/orders" do
    require_token!
    body = json_body
    pet, customer = body["pet_id"], body["customer_id"]
    unless int?(pet) && int?(customer) && DB[:pets].where(id: pet).count == 1 && DB[:customers].where(id: customer).count == 1
      fail!(422, "pet_id and customer_id must name existing rows")
    end
    order = DB[:orders].returning.insert(pet_id: pet, customer_id: customer).first
    status 201
    JSON.generate(order)
  end

  get "/api/orders/:id" do
    require_token!
    id = id_param("id")
    order = id && DB[:orders].where(id: id).first
    fail!(404, "no such order") unless order
    JSON.generate(order)
  end

  not_found { JSON.generate(error: "not found") }
end
