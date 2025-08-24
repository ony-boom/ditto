-- name: GetPackages :many
SELECT
    id,
    name,
    host
FROM
    packages
ORDER BY
    name;

-- name: GetPackageByName :one
SELECT
    id,
    name,
    host
FROM
    packages
WHERE
    name = ?;

-- name: GetPackageByNameAndHost :one
SELECT
    id,
    name,
    host
FROM
    packages
WHERE
    name = ?
    AND (
        host = ?
        OR (
            host IS NULL
            AND ? IS NULL
        )
    );

-- name: CreatePackage :one
INSERT INTO
    packages (name, host)
VALUES
    (?, ?)
RETURNING
    id,
    name,
    host;

-- name: CreatePackageWithoutHost :one
INSERT INTO
    packages (name)
VALUES
    (?)
RETURNING
    id,
    name,
    host;

-- name: UpsertPackage :one
INSERT INTO
    packages (name, host)
VALUES
    (?, ?) ON CONFLICT (name, host) DO
UPDATE
SET
    name = EXCLUDED.name,
    host = EXCLUDED.host
RETURNING
    id,
    name,
    host;

-- name: UpsertPackageWithoutHost :one
INSERT INTO
    packages (name)
VALUES
    (?) ON CONFLICT (name) DO
UPDATE
SET
    name = EXCLUDED.name
RETURNING
    id,
    name,
    host;

-- name: DeletePackage :exec
DELETE FROM
    packages
WHERE
    name = ?;

-- name: DeletePackageByNameAndHost :exec
DELETE FROM
    packages
WHERE
    name = ?
    AND (
        host = ?
        OR (
            host IS NULL
            AND ? IS NULL
        )
    );

-- name: GetPackagesByHost :many
SELECT
    id,
    name,
    host
FROM
    packages
WHERE
    (
        host = ?
        OR (
            host IS NULL
            AND ? IS NULL
        )
    );

-- name: DeletePackagesByHost :exec
DELETE FROM
    packages
WHERE
    (
        host = ?
        OR (
            host IS NULL
            AND ? IS NULL
        )
    );
