defmodule Ark.Repo do
  use Ecto.Repo,
    otp_app: :ark,
    adapter: Ecto.Adapters.Postgres
end
