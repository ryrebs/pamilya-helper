package db

const initSqlStmt = `
CREATE TABLE IF NOT EXISTS account (
	id INTEGER NOT NULL PRIMARY KEY,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	name TEXT,
	email TEXT UNIQUE,
	password TEXT,
	birthdate TEXT NOT NULL DEFAULT '',
	address TEXT NOT NULL DEFAULT '',
	is_verified INTEGER DEFAULT 0,
	is_verification_pending INTEGER DEFAULT 0,
	is_admin INTEGER DEFAULT 0,
	detail TEXT NOT NULL DEFAULT '',
	contact TEXT NOT NULL DEFAULT '',
	profile_image TEXT DEFAULT '',
	gov_id_image TEXT DEFAULT '',
	title TEXT DEFAULT '',
	skills TEXT DEFAULT ''
);
CREATE TABLE IF NOT EXISTS job (
	id INTEGER NOT NULL PRIMARY KEY,
	employer_id INTEGER NOT NULL,
	title TEXT,
	description TEXT,
	responsibility TEXT,
	skills TEXT,
	location TEXT,
	price_from TEXT,
	price_to TEXT,
	employment_type TEXT,
	dateLine TEXT,
	FOREIGN KEY(employer_id) REFERENCES account(id)
);
CREATE TABLE IF NOT EXISTS job_application (
	id INTEGER NOT NULL PRIMARY KEY,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	employee_id INTEGER NOT NULL,
	job_id INTEGER NOT NULL,
	status TEXT DEFAULT 'PENDING',
	FOREIGN KEY(employee_id) REFERENCES account(id),
	FOREIGN KEY(job_id) REFERENCES job(id),
	UNIQUE(employee_id, job_id) ON CONFLICT ROLLBACK
)
`
