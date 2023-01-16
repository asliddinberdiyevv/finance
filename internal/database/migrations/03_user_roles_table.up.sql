-- Create Type Roles to avoid incorrect input
-- We need only 'admin' role
-- Role 'member' is if user exists in database
-- If we will need more roles we will add then to this ENUM
CREATE TYPE user_role AS ENUM ('admin');

CREATE TABLE user_roles (
  user_id UUID NOT NULL REFERENCES users,
  role user_role NOT NULL,
  PRIMARY KEY (user_id, role)
);

-- Create index for roles
CREATE INDEX user_roles_user ON user_roles (user_id)