class PetsController < ApplicationController
  before_action :authenticate!, only: %i[create update]

  def index
    page = integer_param(:page, 1)
    size = integer_param(:page_size, 20)
    filters = { category_id: integer_param(:category), price_min: integer_param(:min_price), price_max: integer_param(:max_price) }
    return render_error(:bad_request, "numeric parameters must be integers") if page == :invalid || size == :invalid || filters.value?(:invalid)
    return render_error(:bad_request, "page must be at least 1 and page_size between 1 and 100") unless page >= 1 && (1..100).cover?(size)

    status = params[:status].presence
    return render_error(:bad_request, "unknown status") if status && !Pet::STATUSES.include?(status)

    pets = Pet.all
    pets = pets.where(category_id: filters[:category_id]) if filters[:category_id]
    pets = pets.where(status: status) if status
    pets = pets.where("price_cents >= ?", filters[:price_min]) if filters[:price_min]
    pets = pets.where("price_cents <= ?", filters[:price_max]) if filters[:price_max]
    render json: { items: pets.order(:id).limit(size).offset((page - 1) * size), page: page, page_size: size, total: pets.count }
  end

  def show
    render json: Pet.find(params[:id])
  end

  def create
    pet = Pet.new(pet_params)
    pet.save ? render(json: pet, status: :created) : render_invalid(pet)
  end

  def update
    pet = Pet.find(params[:id])
    pet.update(pet_params) ? render(json: pet) : render_invalid(pet)
  end

  private

  # An optional integer query parameter: nil when absent, :invalid when it is not an integer in range.
  def integer_param(name, default = nil)
    raw = params[name]
    return default if raw.blank?

    Integer(raw, 10).then { |n| n.abs <= MAX_INT32 ? n : :invalid }
  rescue ArgumentError, TypeError
    :invalid
  end

  def pet_params
    params.permit(:name, :price_cents, :status, :category_id)
  end

  def render_invalid(record)
    render_error :unprocessable_entity, record.errors.full_messages.to_sentence
  end
end
