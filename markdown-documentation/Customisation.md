# Customisation

You may override locales, templates, static content (like CSS) and even database migrations at runtime without
recompiling bbbstatus. Be advised that even though this is possible, it may lead to problems, as only the embedded /
offical versions are tested. You may override multiple files at once. to do that you need to override every file
individually, just as described bellow.

[Back to main documenation](../README.MD)

## Overriding a locale

to override a locale, create / map a folder in the work directory of bbbstatus called `overwrite`, inside this override
folder create a folder named `locales` your setup should look like this:

```
.
├── bbbstatus
└── overwrite
    └── locales
```

I suggest that you now copy every locale you do not want to override into the locales folder. After that make your
modifications / additions, and reload bbbstatus.

## Overriding a template

to override a template, create / map a folder in the work directory of bbbstatus called `overwrite`, inside this
override folder create a folder named `public`, inside the public folder create a folder called `views`.
your setup should look like this:

```
.
├── bbbstatus
└── overwrite
    └── public
        └── views
```

I suggest that you now copy every template you do not want to override into the views folder. After that make your
modifications, and reload bbbstatus.

## Overriding a static file

to override a static file, create / map a folder in the work directory of bbbstatus called `overwrite`, inside this
override folder create a folder named `static` your setup should look like this:

```
.
├── bbbstatus
└── overwrite
    └── static
```

I suggest that you now copy every static file you do not want to override into the `static` folder. After that make your
modifications / additions, and reload bbbstatus.

## Overriding a migration

to override a migration, create / map a folder in the work directory of bbbstatus called `overwrite`, inside this
override folder create a folder named `Migrations`, inside the public folder create a folder called `PostgreSQL`.
your setup should look like this:

```
.
├── bbbstatus
└── overwrite
    └── Migrations
        └── PostgreSQL
```

I suggest that you now copy every migration you do not want to override into the `Migrations` folder. After that make
your modifications / additions, and reload bbbstatus.