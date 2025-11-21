import {
  ProForm,
  ProFormDateTimePicker,
  ProFormSelect,
} from "@ant-design/pro-components";
import { Space } from "antd";
import React from "react";
import type { QueryEventDataRequest } from "../../types";

interface EventDataQueryFormProps {
  form: any;
  eventConfigs: { label: string; value: string }[];
  onSearch: (values: QueryEventDataRequest) => void;
}

const EventDataQueryForm: React.FC<EventDataQueryFormProps> = ({
  form,
  eventConfigs,
  onSearch,
}) => {
  return (
    <ProForm
      form={form}
      layout="horizontal"
      onFinish={onSearch}
      submitter={{
        searchConfig: {
          submitText: "查询",
          resetText: "重置",
        },
        render: (_, dom) => <Space>{dom}</Space>,
      }}
    >
      <ProFormSelect
        name="eventId"
        label="事件名称"
        options={eventConfigs}
        placeholder="请选择事件"
        rules={[{ required: true, message: "请选择一个事件进行查询" }]}
        showSearch
      />
      <ProFormDateTimePicker name="startTime" label="开始时间" />
      <ProFormDateTimePicker name="endTime" label="结束时间" />
    </ProForm>
  );
};

export default EventDataQueryForm;
