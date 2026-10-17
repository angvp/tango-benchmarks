class SessionsController < ApplicationController
  def create
    staff = StaffMember.find_by(username: params[:username].to_s)
    if staff&.authenticate(params[:password].to_s)
      render json: { token: issue_token }
    else
      render_error :unauthorized, "bad credentials"
    end
  end
end
