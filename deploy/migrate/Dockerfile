FROM migrate/migrate:v4.19.1

# Bake SQL migrations into the image.
# This allows running `migrate` on the server without mounting the repository code.
COPY backend/migrations /migrations

