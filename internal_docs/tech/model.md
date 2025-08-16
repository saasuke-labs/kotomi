# Data Model

## Owner

An `Owner` is a registered user that uses `Kotomi` to add dynamic content to their sites.

| field | type   | description                  | validation           |
| ----- | ------ | ---------------------------- | -------------------- |
| id    | string | the identifier for the owner | uuid                 |
| name  | string | the name of the owner        | minLen=5, maxLen=100 |

## Sites

A `Site` is an application that is available in one or more domains and it is owned and managed by an `Owner`.

| field   | type   | description                 | format               |
| ------- | ------ | --------------------------- | -------------------- |
| id      | uuid   | the identifier for the site |                      |
| name    | string | the name of the site        | minLen=5, maxLen=100 |
| ownerId | string | the owner of the site       | uuid                 |

## Pages

Name TBD

A `Page` is a specific part of the site the comment refers to. A `Page` belongs to a `Site`

| field  | type   | description                       | format     |
| ------ | ------ | --------------------------------- | ---------- |
| slug   | string | the slug / identifier of the page | maxLen=200 |
| siteId | string | the site the page belongs to      | uuid       |

## Comments

A comment is a text that is posted for a given `Site` and a specific `Page`.

| field       | type     | description                            | format                |
| ----------- | -------- | -------------------------------------- | --------------------- |
| id          | string   | the id of the comment                  | uuid                  |
| slug        | string   | the slug / identifier of the page      | maxLen=200            |
| siteId      | string   | the site the page belongs to           | uuid                  |
| text        | string   | the content of the comment             | minLen=5, maxLen=2000 |
| postedAt    | datetime | the moment the comment was posted      |                       |
| publishedAt | datetime | the moment the comment was made public |                       |
