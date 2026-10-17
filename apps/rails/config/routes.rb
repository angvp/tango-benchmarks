Rails.application.routes.draw do
  scope :api do
    post "login", to: "sessions#create"
    resources :pets, only: %i[index show create update]
    resources :orders, only: %i[create show]
  end
end
