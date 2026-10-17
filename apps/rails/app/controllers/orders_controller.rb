class OrdersController < ApplicationController
  before_action :authenticate!

  def create
    order = Order.new(params.permit(:pet_id, :customer_id))
    order.save ? render(json: order, status: :created) : render_error(:unprocessable_entity, "pet_id and customer_id must name existing rows")
  end

  def show
    render json: Order.find(params[:id])
  end
end
