
CREATE TABLE titan.scrubs (
    manu text not null,
    scid bigint not null,
    CONSTRAINT PRIMARY KEY (manu, scid)
);

CREATE TABLE titan.rebates (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    stat text not null default '',
    fprt text not null default '',
    CONSTRAINT PRIMARY KEY (manu, scid, rbid)
);
CREATE INDEX ON titan.rebates(manu);
CREATE INDEX ON titan.rebates(manu, scid);

CREATE TABLE titan.claims (
    manu text not null,
    scid bigint not null,
    shrt text not null,
    excl text not null default '',
    CONSTRAINT PRIMARY KEY (manu, scid, shrt)
);
CREATE INDEX ON titan.claims(manu, scid);
CREATE INDEX ON titan.claims(manu, shrt);
CREATE INDEX ON titan.claims(manu, scid, shrt);

CREATE TABLE titan.rebate_claims (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    shrt text not null,
    CONSTRAINT PRIMARY KEY (manu, scid, rbid, shrt)
);