CREATE TABLE IF NOT EXISTS "users"
(
    id       SERIAL PRIMARY KEY,
    email    VARCHAR(255),
    password VARCHAR(255),
    username VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS "roles"
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS "userRoles"
(
    id     SERIAL PRIMARY KEY,
    roleId INT NOT NULL references roles(id),
    userId INT NOT NULL references users(id)
)
