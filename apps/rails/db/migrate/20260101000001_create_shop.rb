class CreateShop < ActiveRecord::Migration[8.1]
  def change
    create_table :categories do |t|
      t.string :name, null: false
    end

    create_table :pets do |t|
      t.string :name, null: false
      t.integer :price_cents, null: false
      t.string :status, null: false
      t.references :category, null: false, foreign_key: true, type: :integer
    end
    add_index :pets, :status

    create_table :customers do |t|
      t.string :name, null: false
      t.string :email, null: false
    end

    create_table :orders do |t|
      t.references :pet, null: false, foreign_key: true, type: :integer
      t.references :customer, null: false, foreign_key: true, type: :integer
    end

    create_table :staff_members do |t|
      t.string :username, null: false, index: { unique: true }
      t.string :password_digest, null: false
    end
  end
end
