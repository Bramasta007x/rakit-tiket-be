INSERT INTO "user" (id, name, email, password_hash, role, created_at, updated_at)
VALUES
(
    gen_random_uuid(), 
    'Super Admin',
    'admin@rakittiket.com',
    -- Hash bcrypt untuk password: "rakit123"
    '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW', 
    'ADMIN',
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'Petugas Lapangan',
    'staff@rakittiket.com',
    -- Hash bcrypt untuk password: "rakit123"
    '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW', 
    'GROUND STAFF',
    NOW(),
    NOW()
);