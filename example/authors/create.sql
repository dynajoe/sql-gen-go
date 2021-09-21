INSERT INTO authors (
  name, bio
) VALUES (
  :name, :bio
)
RETURNING *;