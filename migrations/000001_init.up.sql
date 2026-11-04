CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_lowercase CHECK (email = lower(email)),
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 120),
    slug TEXT NOT NULL UNIQUE,
    default_currency CHAR(3) NOT NULL DEFAULT 'EUR',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organization_members (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'manager', 'staff', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, user_id)
);
CREATE INDEX organization_members_user_idx ON organization_members(user_id);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);
CREATE INDEX sessions_user_active_idx ON sessions(user_id, expires_at)
    WHERE revoked_at IS NULL;

CREATE TABLE customers (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 160),
    email TEXT,
    phone TEXT,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, id),
    CONSTRAINT customers_email_nonempty CHECK (email IS NULL OR length(trim(email)) > 0)
);
CREATE INDEX customers_org_created_idx ON customers(organization_id, created_at DESC);
CREATE INDEX customers_org_email_idx ON customers(organization_id, email) WHERE email IS NOT NULL;

CREATE TABLE products (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 200),
    description TEXT NOT NULL DEFAULT '',
    price_minor BIGINT NOT NULL CHECK (price_minor >= 0),
    currency CHAR(3) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, id)
);
CREATE INDEX products_org_active_idx ON products(organization_id, is_active);

CREATE TABLE orders (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id UUID,
    status TEXT NOT NULL CHECK (status IN ('draft', 'pending', 'confirmed', 'completed', 'cancelled', 'refunded')),
    total_minor BIGINT NOT NULL CHECK (total_minor >= 0),
    currency CHAR(3) NOT NULL,
    idempotency_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, id),
    CONSTRAINT orders_customer_same_org_fk
        FOREIGN KEY (organization_id, customer_id)
        REFERENCES customers(organization_id, id),
    CONSTRAINT orders_idempotency_unique UNIQUE (organization_id, idempotency_key)
);
CREATE INDEX orders_org_created_idx ON orders(organization_id, created_at DESC);
CREATE INDEX orders_org_status_idx ON orders(organization_id, status, created_at DESC);

CREATE TABLE order_items (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    order_id UUID NOT NULL,
    product_id UUID,
    product_name_snapshot TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_minor BIGINT NOT NULL CHECK (unit_price_minor >= 0),
    currency CHAR(3) NOT NULL,
    PRIMARY KEY (organization_id, id),
    CONSTRAINT order_items_order_same_org_fk
        FOREIGN KEY (organization_id, order_id)
        REFERENCES orders(organization_id, id) ON DELETE CASCADE,
    CONSTRAINT order_items_product_same_org_fk
        FOREIGN KEY (organization_id, product_id)
        REFERENCES products(organization_id, id)
);
CREATE INDEX order_items_order_idx ON order_items(organization_id, order_id);

CREATE TABLE bot_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    telegram_bot_id BIGINT NOT NULL,
    encrypted_token BYTEA NOT NULL,
    webhook_secret_hash BYTEA NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, telegram_bot_id),
    UNIQUE (id, organization_id)
);
CREATE INDEX bot_instances_org_active_idx ON bot_instances(organization_id, is_active);

CREATE TABLE processed_telegram_updates (
    bot_instance_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    update_id BIGINT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (bot_instance_id, update_id),
    CONSTRAINT processed_updates_bot_same_org_fk
        FOREIGN KEY (bot_instance_id, organization_id)
        REFERENCES bot_instances(id, organization_id) ON DELETE CASCADE
);

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    schema_version INTEGER NOT NULL DEFAULT 1 CHECK (schema_version > 0),
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    last_error TEXT
);
CREATE INDEX outbox_pending_idx ON outbox_events(available_at, occurred_at)
    WHERE published_at IS NULL;

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX audit_logs_org_created_idx ON audit_logs(organization_id, created_at DESC);
