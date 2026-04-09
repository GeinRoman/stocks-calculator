ALTER TABLE profiles ADD CONSTRAINT profiles_user_id_name_key UNIQUE (user_id, name);
