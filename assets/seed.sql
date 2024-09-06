CREATE TABLE if not exists "resource" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE,
	"template"	TEXT,
	"generator"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "person_prop" (
	"id"	INTEGER NOT NULL UNIQUE,
	"type"	TEXT NOT NULL,
	"value"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);

insert or ignore into resource values 
	(null, "person.firstName", null, "random_row:person_prop:type=firstName"),
	(null, "person.middleName", null, "random_row:person_prop:type=middleName"),
	(null, "person.lastName", null, "random_row:person_prop:type=lastName"),
	(null, "person.nickName", null, "random_row:person_prop:type=nickName"),
	(null, "person.age", "0..100", "int_range"),
	(null, "person.phone", null, "random_row:person_prop:type=nickName"),
	(null, "person.email", "{firstName}.{lastName}@{provider},{lastName}.{firstName}@{provider},{nickName}@provider", "email"),
	(null, "location.country", null, "random_row:location_prop:type=country"),
	(null, "location.town", null, "random_row:location_prop:type=town"),
	(null, "location.continent", null, "random_row:location_prop:type=continent")
    ;
	