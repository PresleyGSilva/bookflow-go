-- BookFlow seed: Barbearia Corte Fino (demo data, future appointments relative to now)
DO $$
DECLARE
  v_org UUID;
  v_user UUID;
  v_p1 UUID; v_p2 UUID; v_p3 UUID;
  v_s1 UUID; v_s2 UUID; v_s3 UUID; v_s4 UUID;
  v_c1 UUID; v_c2 UUID;
  wd INT;
BEGIN
  -- idempotent: skip when the demo admin already exists
  IF EXISTS (SELECT 1 FROM users WHERE email='admin@cortefino.com') THEN RETURN; END IF;
  INSERT INTO orgs(name) VALUES ('Barbearia Corte Fino') RETURNING id INTO v_org;
  INSERT INTO users(name,email,password_hash)
    VALUES ('Admin','admin@cortefino.com','$2b$10$0jadwsjUDMbF3pMKtGh7hOt7tAUSYmFLRlUFvtV9jLcb8lEr.ko52')
    RETURNING id INTO v_user; -- password: admin123
  INSERT INTO org_members(org_id,user_id,role) VALUES (v_org,v_user,'admin');
  INSERT INTO settings(org_id) VALUES (v_org);

  INSERT INTO professionals(org_id,name,email,phone) VALUES
    (v_org,'Carlos Silva','carlos@cortefino.com','(11) 99911-1111'),
    (v_org,'João Pedro','joao@cortefino.com','(11) 99922-2222'),
    (v_org,'Rafael Souza','rafael@cortefino.com','(11) 99933-3333');

  SELECT id INTO v_p1 FROM professionals WHERE org_id=v_org AND name='Carlos Silva';
  SELECT id INTO v_p2 FROM professionals WHERE org_id=v_org AND name='João Pedro';
  SELECT id INTO v_p3 FROM professionals WHERE org_id=v_org AND name='Rafael Souza';

  INSERT INTO services(org_id,name,duration_min,price_cents) VALUES
    (v_org,'Corte Clássico',30,4500),
    (v_org,'Barba',20,3000),
    (v_org,'Corte + Barba',50,7000),
    (v_org,'Coloração',60,9000);

  SELECT id INTO v_s1 FROM services WHERE org_id=v_org AND name='Corte Clássico';
  SELECT id INTO v_s2 FROM services WHERE org_id=v_org AND name='Barba';

  -- working hours Mon(1)-Sat(6): 08:00-12:00 + 13:00-19:00
  FOR wd IN 1..6 LOOP
    INSERT INTO working_hours(org_id,professional_id,weekday,start_min,end_min)
    VALUES (v_org,v_p1,wd,480,720),(v_org,v_p1,wd,780,1140),
           (v_org,v_p2,wd,480,720),(v_org,v_p2,wd,780,1140),
           (v_org,v_p3,wd,480,720),(v_org,v_p3,wd,780,1140);
  END LOOP;

  INSERT INTO customers(org_id,name,email,phone) VALUES
    (v_org,'Marcos Oliveira','marcos@mail.com','(11) 98888-1111'),
    (v_org,'Paulo Santos','paulo@mail.com','(11) 98888-2222');

  SELECT id INTO v_c1 FROM customers WHERE org_id=v_org AND name='Marcos Oliveira';
  SELECT id INTO v_c2 FROM customers WHERE org_id=v_org AND name='Paulo Santos';

  -- a few future appointments (tomorrow 09:00 and 10:00 São Paulo time)
  INSERT INTO appointments(org_id,professional_id,service_id,customer_id,customer_name,code,starts_at,ends_at)
  VALUES
    (v_org,v_p1,v_s1,v_c1,'Marcos Oliveira','BF-SEED01',
      (now() + interval '1 day')::date + time '09:00' AT TIME ZONE 'America/Sao_Paulo',
      (now() + interval '1 day')::date + time '09:30' AT TIME ZONE 'America/Sao_Paulo'),
    (v_org,v_p2,v_s2,v_c2,'Paulo Santos','BF-SEED02',
      (now() + interval '1 day')::date + time '10:00' AT TIME ZONE 'America/Sao_Paulo',
      (now() + interval '1 day')::date + time '10:20' AT TIME ZONE 'America/Sao_Paulo');
END $$;
