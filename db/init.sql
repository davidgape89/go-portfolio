CREATE TABLE users(
   id serial PRIMARY KEY,
   username VARCHAR (50) UNIQUE NOT NULL,
   password VARCHAR (80) NOT NULL,
   email VARCHAR (355) UNIQUE NOT NULL,
   created_on TIMESTAMP NOT NULL,
   last_login TIMESTAMP
);

CREATE TYPE post_status AS ENUM ('posted', 'hidden');

CREATE TABLE posts(
   	id serial PRIMARY KEY,
	user_id INTEGER NOT NULL,
	title VARCHAR (255) NOT NULL,
	content TEXT NOT NULL,
	status post_status NOT NULL,
	create_time TIMESTAMP NOT NULL,
	update_time TIMESTAMP
);