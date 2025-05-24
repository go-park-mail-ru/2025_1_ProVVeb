CREATE OR REPLACE FUNCTION translit(text) RETURNS text AS $$
SELECT translate($1,
    'АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯабвгдеёжзийклмнопрстуфхцчшщъыьэюя',
    'ABVGDEEJZIJKLMNOPRSTUFHCCHSHSH_Y_EUAabvgdeejzijklmnoprstufhcchshsh_y_eua'
);
$$ LANGUAGE SQL IMMUTABLE;

CREATE OR REPLACE FUNCTION update_fullname_translit()
RETURNS trigger AS $$
BEGIN
  NEW.fullname_translit := translit(NEW.firstname || ' ' || NEW.lastname);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_fullname_translit
BEFORE INSERT OR UPDATE OF firstname, lastname
ON profiles
FOR EACH ROW
EXECUTE FUNCTION update_fullname_translit();


CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER set_updated_at
BEFORE UPDATE ON profiles
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER set_updated_at_on_users
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER set_updated_at_on_static
BEFORE UPDATE ON static
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER set_updated_at_on_messahes
BEFORE UPDATE ON messages
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();