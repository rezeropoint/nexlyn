-- 初始化脚本：创建附加数据库
-- 此脚本会在容器第一次初始化时由 Postgres 执行

-- 创建数据库 
CREATE DATABASE mediahandler WITH ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0;
