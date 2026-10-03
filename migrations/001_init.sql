-- IMS Finance - Initial Migration
-- Schema: ujicoba

-- Create schema
CREATE SCHEMA IF NOT EXISTS ujicoba;

-- Table: clients
CREATE TABLE IF NOT EXISTS ujicoba.clients (
    id         BIGSERIAL PRIMARY KEY,
    client_code VARCHAR(50)  NOT NULL UNIQUE,
    name        VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Table: contracts
CREATE TABLE IF NOT EXISTS ujicoba.contracts (
    id                  BIGSERIAL PRIMARY KEY,
    contract_no         VARCHAR(50)    NOT NULL UNIQUE,
    client_id           BIGINT         NOT NULL REFERENCES ujicoba.clients(id),
    otr                 NUMERIC(18,2)  NOT NULL,
    dp_percentage       NUMERIC(5,2)   NOT NULL,
    dp_amount           NUMERIC(18,2)  NOT NULL,
    principal_amount    NUMERIC(18,2)  NOT NULL,
    tenor_months        INTEGER        NOT NULL,
    interest_rate       NUMERIC(7,4)   NOT NULL,
    interest_amount     NUMERIC(18,2)  NOT NULL,
    total_amount        NUMERIC(18,2)  NOT NULL,
    monthly_installment NUMERIC(18,2)  NOT NULL,
    contract_date       DATE           NOT NULL,
    first_due_date      DATE           NOT NULL,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CHECK (otr > 0),
    CHECK (dp_percentage >= 0 AND dp_percentage <= 100),
    CHECK (tenor_months > 0),
    CHECK (interest_rate >= 0)
);

-- Table: installments
CREATE TABLE IF NOT EXISTS ujicoba.installments (
    id               BIGSERIAL PRIMARY KEY,
    contract_id      BIGINT        NOT NULL REFERENCES ujicoba.contracts(id),
    installment_no   INTEGER       NOT NULL,
    due_date         DATE          NOT NULL,
    amount           NUMERIC(18,2) NOT NULL,
    principal_amount NUMERIC(18,2),
    interest_amount  NUMERIC(18,2),
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CHECK (installment_no > 0),
    CHECK (amount > 0),
    UNIQUE (contract_id, installment_no)
);

-- Table: payments
CREATE TABLE IF NOT EXISTS ujicoba.payments (
    id                BIGSERIAL PRIMARY KEY,
    installment_id    BIGINT        NOT NULL REFERENCES ujicoba.installments(id),
    payment_date      DATE          NOT NULL,
    amount            NUMERIC(18,2) NOT NULL,
    payment_reference VARCHAR(100),
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CHECK (amount > 0)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_contracts_client_id
    ON ujicoba.contracts(client_id);

CREATE INDEX IF NOT EXISTS idx_installments_contract_due_date
    ON ujicoba.installments(contract_id, due_date);

CREATE INDEX IF NOT EXISTS idx_payments_installment_id
    ON ujicoba.payments(installment_id);
