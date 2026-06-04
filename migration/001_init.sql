DROP DATABASE values_db;
CREATE DATABASE values_db OWNER root;

\c values_db

CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	name TEXT,
	role TEXT
);

CREATE TABLE vals (
	id SERIAL PRIMARY KEY,
	data INTEGER
);
