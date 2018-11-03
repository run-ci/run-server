CREATE TABLE git_repos (
    remote varchar(255) NOT NULL UNIQUE
);

INSERT INTO git_repos (remote)
VALUES
    ('https://github.com/run-ci/run-server.git');
