CREATE TABLE `service` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '英文标识名称',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
  `dir` varchar(255) NOT NULL DEFAULT '' COMMENT '目录',
  `cmd_start` varchar(500) NOT NULL DEFAULT '' COMMENT '启动脚本',
  `cmd_stop` varchar(500) NOT NULL DEFAULT '' COMMENT '关闭脚本',
  `cmd_restart` varchar(500) NOT NULL DEFAULT '' COMMENT '重启脚本',
  `port` int(11) NOT NULL DEFAULT 0 COMMENT '端口',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_at` datetime NOT NULL DEFAULT current_timestamp() COMMENT '添加时间',
  `update_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_general_ci;