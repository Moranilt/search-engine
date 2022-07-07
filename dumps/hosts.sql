CREATE TABLE hosts (
  id SERIAL PRIMARY KEY,
  name VARCHAR UNIQUE,
  is_searchable BOOLEAN,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE phrases (
  id SERIAL PRIMARY KEY,
  name VARCHAR UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE endpoints (
  id SERIAL PRIMARY KEY,
  host_id INT NOT NULL,
  name VARCHAR NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (host_id) REFERENCES hosts (id) ON DELETE CASCADE
);

CREATE TABLE titles (
	id SERIAL PRIMARY KEY,
	endpoint_id INT NOT NULL,
	value VARCHAR NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	FOREIGN KEY (endpoint_id) REFERENCES endpoints (id) ON DELETE CASCADE
);

CREATE TABLE endpoints_phrases (
  phrase_id INT,
  endpoint_id INT,
  FOREIGN KEY (phrase_id) REFERENCES phrases (id) ON DELETE CASCADE,
  FOREIGN KEY (endpoint_id) REFERENCES endpoints (id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION create_phrase("search_phrase" text)
	RETURNS integer
	LANGUAGE plpgsql
AS
$BODY$
	DECLARE phrase_id integer;
	BEGIN
		if EXISTS (SELECT FROM phrases WHERE name=search_phrase) THEN
			SELECT id INTO phrase_id FROM phrases WHERE name=search_phrase;
		ELSE
			INSERT INTO phrases (name) VALUES (search_phrase) RETURNING id INTO phrase_id;
		END IF;
    return phrase_id;
	END;
$BODY$;

CREATE OR REPLACE FUNCTION create_endpoint("host" integer, "endpoint" text)
	RETURNS integer
	LANGUAGE plpgsql
AS
$BODY$
	DECLARE endpoint_id integer;
	BEGIN
		if EXISTS (SELECT FROM endpoints WHERE host_id=host AND name=endpoint) THEN
			SELECT id INTO endpoint_id 
        FROM endpoints 
        WHERE host_id=host AND name=endpoint;
		ELSE
			INSERT INTO 
      endpoints (host_id, name) 
      VALUES(host, endpoint) 
      RETURNING id INTO endpoint_id;
		END IF;
		return endpoint_id;
	END;
$BODY$;

CREATE OR REPLACE FUNCTION create_title("endpoint" integer, "title" text)
	RETURNS void
	LANGUAGE plpgsql
AS
$BODY$
	BEGIN
		IF NOT EXISTS (SELECT FROM titles WHERE endpoint_id=endpoint AND value=title) THEN
			INSERT INTO titles (endpoint_id, value) VALUES (endpoint, title);
		END IF;
	END;
$BODY$;

CREATE OR REPLACE FUNCTION create_endpoint_title("host" integer, "endpoint" text, "title" text)
	RETURNS integer
	LANGUAGE plpgsql
AS
$BODY$
	DECLARE endpoint_id integer;
	BEGIN
		SELECT create_endpoint(host, endpoint) INTO endpoint_id;
		PERFORM create_title(endpoint_id, title);
		return endpoint_id;
	END;
$BODY$;

CREATE OR REPLACE FUNCTION create_endpoint_phrase_title("host" integer, "endpoint" text, "search_phrase" text, "title" text)
	returns void
	LANGUAGE plpgsql
AS
$BODY$
	DECLARE 
		last_phrase_id integer;
		last_endpoint_id integer;
	BEGIN
		SELECT create_phrase(search_phrase) INTO last_phrase_id;
		SELECT create_endpoint_title(host, endpoint, title) INTO last_endpoint_id;
    IF NOT EXISTS (SELECT FROM endpoints_phrases WHERE phrase_id=last_phrase_id AND endpoint_id=last_endpoint_id) THEN
			INSERT INTO endpoints_phrases (phrase_id, endpoint_id) VALUES (last_phrase_id, last_endpoint_id);
		END IF;
	END;
$BODY$;

INSERT INTO hosts (name, is_searchable) VALUES ('https://www.spacex.com/', true);