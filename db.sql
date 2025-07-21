-- 服务管理工具数据库脚本
CREATE TABLE `service` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '英文标识名称',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
  `dir` varchar(255) NOT NULL DEFAULT '' COMMENT '目录',
  `cmd_start` varchar(500) NOT NULL DEFAULT '' COMMENT '启动脚本',
  `cmd_stop` varchar(500) NOT NULL DEFAULT '' COMMENT '关闭脚本',
  `cmd_restart` varchar(500) NOT NULL DEFAULT '' COMMENT '重启脚本',
  `port` int(11) NOT NULL DEFAULT 0 COMMENT '端口',
  `health_check_url` varchar(500) DEFAULT '' COMMENT '健康检查URL',
  `auto_restart` tinyint(1) DEFAULT 0 COMMENT '是否自动重启',
  `max_restart_count` int(11) DEFAULT 3 COMMENT '最大重启次数',
  `restart_interval` int(11) DEFAULT 30 COMMENT '重启间隔(秒)',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT current_timestamp() COMMENT '添加时间',
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb3;

-- 创建服务日志表
CREATE TABLE IF NOT EXISTS `service_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `service_id` bigint(20) NOT NULL COMMENT '服务ID',
  `operation` varchar(50) NOT NULL COMMENT '操作类型',
  `status` varchar(20) NOT NULL COMMENT '操作状态',
  `output` text COMMENT '操作输出',
  `error` text COMMENT '错误信息',
  `duration` bigint(20) DEFAULT 0 COMMENT '执行时长(毫秒)',
  `created_at` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_service_id` (`service_id`),
  KEY `idx_operation` (`operation`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务操作日志表';
