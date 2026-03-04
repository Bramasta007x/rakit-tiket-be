INSERT INTO "user" (id, name, email, password_hash, role, created_at, updated_at)
VALUES
(
    gen_random_uuid(), 
    'Super Admin',
    'admin@rakittiket.com',
    -- Hash bcrypt untuk password: "rakit123"
    '$2a$12$uaTYijRNN03elRWt4YhVQ.QMkVpGceiIugR8Dkz/0lTcptaxIxxEK', 
    'ADMIN',
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'Petugas Lapangan',
    'staff@rakittiket.com',
    -- Hash bcrypt untuk password: "rakit123"
    '$2a$12$uaTYijRNN03elRWt4YhVQ.QMkVpGceiIugR8Dkz/0lTcptaxIxxEK', 
    'GROUND STAFF',
    NOW(),
    NOW()
);