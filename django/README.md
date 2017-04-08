# django

We will eventually open source this repo on github.com/acko/go-django.

This library provides interoperability with django's sessions and user
authentication system.

The library is organised as `django` package that gives high level interface, and
`django/backend/db`, `django/backend/redis` etc.

User and Session stores can be used with different backends within same project.
