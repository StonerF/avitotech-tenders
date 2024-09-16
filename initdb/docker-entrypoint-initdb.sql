CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE TABLE employee (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
);

CREATE TABLE organization (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type organization_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_responsible (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE
);

CREATE TYPE servicetype AS ENUM (
	    'Construction',
	    'Delivery',
	    'Manufacture'
	);

CREATE TYPE status_type AS ENUM (
	    'Created',
	    'Published',
	    'Closed'
	);

CREATE TABLE IF NOT EXISTS tenders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
	description TEXT NOT NULL,
    type servicetype,
    status  status_type,
    organizationid UUID,
    creatorusername VARCHAR(50),
	version INT,
	createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE authortype AS ENUM(
	    'Organization',
	    'User'
	);
CREATE TYPE status_typebid AS ENUM(
	    'Created',
	    'Published',
	    'Canceled'
	);

CREATE TABLE IF NOT EXISTS bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
	description TEXT NOT NULL,
    status  status_typebid,
	tenderid UUID,
	type authortype,
    authorid UUID,
	version INT,
	createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);