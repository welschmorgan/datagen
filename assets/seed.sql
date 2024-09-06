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
	(null, "person.firstName", "person_prop:type=firstName", "random_row"),
	(null, "person.middleName", "person_prop:type=middleName", "random_row"),
	(null, "person.lastName", "person_prop:type=lastName", "random_row"),
	(null, "person.nickName", "person_prop:type=nickName", "random_row"),
	(null, "person.age", "0..100", "int_range"),
	(null, "person.phone", "person_prop:type=nickName", "random_row"),
	(null, "person.email", "{firstName}.{lastName}@{provider},{lastName}.{firstName}@{provider},{nickName}@provider", "email"),
	(null, "location.country", "location_prop:type=country", "random_row"),
	(null, "location.town", "location_prop:type=town", "random_row"),
	(null, "location.continent", "location_prop:type=continent", "random_row")
    ;
	