# Исправление прав доступа к базе данных

Если вы получаете ошибку `permission denied for table users`, выполните следующие шаги:

## Вариант 1: Выполнить SQL скрипт напрямую

Подключитесь к базе данных как суперпользователь (обычно `postgres`) и выполните:

```bash
psql -U postgres -d crowdsourcing_db -f scripts/fix_permissions.sql
```

Или вручную:

```sql
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO crowdsourcing;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO crowdsourcing;
GRANT USAGE, CREATE ON SCHEMA public TO crowdsourcing;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO crowdsourcing;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO crowdsourcing;
```

## Вариант 2: Пересоздать базу данных

Если вы используете Docker Compose:

```bash
docker-compose down -v
docker-compose up -d
```

Это удалит все данные и пересоздаст базу с правильными правами.

## Вариант 3: Выполнить через psql

```bash
psql -U postgres -d crowdsourcing_db
```

Затем выполните команды из `scripts/fix_permissions.sql`

## Проверка прав

После выполнения скрипта проверьте права:

```sql
\dp users
```

Должны быть права для пользователя `crowdsourcing`.

