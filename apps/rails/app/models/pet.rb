class Pet < ApplicationRecord
  STATUSES = %w[available pending sold].freeze

  belongs_to :category

  validates :name, presence: true
  validates :price_cents, numericality: { only_integer: true, greater_than_or_equal_to: 0, less_than_or_equal_to: 2_147_483_647 }
  validates :status, inclusion: { in: STATUSES, message: "must be available, pending or sold" }

  def as_json(*) = slice(:id, :name, :price_cents, :status, :category_id)
end
