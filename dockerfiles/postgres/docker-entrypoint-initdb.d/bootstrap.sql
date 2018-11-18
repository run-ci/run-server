CREATE TABLE git_repos (
    remote varchar(255) NOT NULL,
    branch varchar(255) NOT NULL,

    PRIMARY KEY(remote, branch),
    UNIQUE(remote, branch)
);

INSERT INTO git_repos (remote, branch)
VALUES
    ('https://github.com/run-ci/run-server.git', 'master');
