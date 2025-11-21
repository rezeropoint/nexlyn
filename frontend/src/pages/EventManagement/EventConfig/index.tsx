import { deleteEventConfig } from "@/services/eventhandler";
import { type ActionType, PageContainer } from "@ant-design/pro-components";
import { message } from "antd";
import React, { useRef, useState } from "react";
import { MESSAGE } from "../constants";
import type { EventConfigWithFields } from "../types";
import EventConfigTable from "./components/EventConfigTable";
import EventConfigForm from "./EventConfigForm"; // 使用正确的表单组件

const EventConfig: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [formOpen, setFormOpen] = useState(false);
  const [currentConfig, setCurrentConfig] = useState<
    EventConfigWithFields | undefined
  >();

  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  const handleCreate = () => {
    setCurrentConfig(undefined);
    setFormOpen(true);
  };

  const handleEdit = (record: EventConfigWithFields) => {
    setCurrentConfig(record);
    setFormOpen(true);
  };

  const handleDelete = async (record: EventConfigWithFields) => {
    try {
      const response = await deleteEventConfig(record.id);
      if (response.code === 0) {
        message.success(MESSAGE.DELETE_SUCCESS);
        handleTableRefresh();
        return true;
      }
      message.error(response.msg || MESSAGE.DELETE_FAILED);
      return false;
    } catch (_error) {
      message.error(MESSAGE.DELETE_FAILED);
      return false;
    }
  };

  const handleSuccess = () => {
    setFormOpen(false);
    setCurrentConfig(undefined);
    handleTableRefresh();
  };

  return (
    <PageContainer subTitle="配置Skylark流程事件的显示字段和组织字段映射">
      <EventConfigTable
        actionRef={actionRef}
        onCreate={handleCreate}
        onEdit={handleEdit}
        onDelete={handleDelete}
      />
      <EventConfigForm
        open={formOpen}
        onOpenChange={setFormOpen}
        currentConfig={currentConfig}
        onSuccess={handleSuccess}
      />
    </PageContainer>
  );
};

export default EventConfig;
