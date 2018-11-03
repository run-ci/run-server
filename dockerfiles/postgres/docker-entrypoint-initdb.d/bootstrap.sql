CREATE TABLE git_triggers (
    remote varchar(255) NOT NULL UNIQUE
);

INSERT INTO git_triggers (remote)
VALUES
    ('https://github.com/run-ci/run');
