class Order < ApplicationRecord
  belongs_to :pet
  belongs_to :customer

  def as_json(*) = slice(:id, :pet_id, :customer_id)
end
