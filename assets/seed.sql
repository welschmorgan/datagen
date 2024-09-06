CREATE TABLE if not exists "resource" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE,
	"generator"	TEXT,
	"template"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "locale" (
	"id" INTEGER NOT NULL UNIQUE,
	"name" TEXT NOT NULL UNIQUE,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "person_prop" (
	"id"	INTEGER NOT NULL UNIQUE,
	"locale_id" INTEGER NOT NULL,
	"type"	TEXT NOT NULL,
	"value"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);

insert or ignore into locale values 
	(null, "fr-FR"),
	(null, "en-UK"),
	(null, "en-US"),
	(null, "es-ES")
	;

insert or ignore into resource values 
	(null, "person.firstName", "random_row", "person_prop:type=firstName"),
	(null, "person.middleName", "random_row", "person_prop:type=middleName"),
	(null, "person.lastName", "random_row", "person_prop:type=lastName"),
	(null, "person.nickName", "random_row", "person_prop:type=nickName"),
	(null, "person.age", "int_range", "0..100"),
	(null, "person.phone", "random_row", "person_prop:type=nickName"),
	(null, "person.email", "email", "{firstName}.{lastName}@{provider},{lastName}.{firstName}@{provider},{nickName}@provider"),
	(null, "location.country", "random_row", "location_prop:type=country"),
	(null, "location.town", "random_row", "location_prop:type=town"),
	(null, "location.continent", "random_row", "location_prop:type=continent")
    ;
	