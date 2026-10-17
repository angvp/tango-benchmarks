class ApplicationController < ActionController::API
  TOKEN_LIFETIME = 12.hours
  MAX_INT32 = 2_147_483_647

  rescue_from ActiveRecord::RecordNotFound do
    render_error :not_found, "not found"
  end
  rescue_from ActionDispatch::Http::Parameters::ParseError do
    render_error :bad_request, "body is not valid JSON"
  end

  private

  def render_error(status, message)
    render json: { error: message }, status: status
  end

  def token_verifier = Rails.application.message_verifier(:staff)

  def issue_token = token_verifier.generate("staff", expires_in: TOKEN_LIFETIME)

  def authenticate!
    token = request.authorization.to_s.delete_prefix("Bearer ")
    render_error(:unauthorized, "a valid bearer token is required") unless token_verifier.verified(token)
  end
end
