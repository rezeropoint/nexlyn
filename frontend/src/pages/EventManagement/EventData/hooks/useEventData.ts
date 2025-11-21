import {
  getEventConfigList,
  getEventDetail,
  queryEventData,
} from "@/services/eventhandler";
import type { ProColumns } from "@ant-design/pro-components";
import { message } from "antd";
import { useCallback, useEffect, useState } from "react";
import type { EventConfigWithFields, QueryEventDataRequest } from "../../types";

export const useEventData = () => {
  const [eventConfigs, setEventConfigs] = useState<
    { label: string; value: string }[]
  >([]);
  const [allEventConfigs, setAllEventConfigs] = useState<
    EventConfigWithFields[]
  >([]);
  const [tableColumns, setTableColumns] = useState<ProColumns<any>[]>([]);
  const [tableData, setTableData] = useState<any[]>([]);
  const [tableLoading, setTableLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10 });
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [eventDetail, setEventDetail] = useState<any>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [currentEventId, setCurrentEventId] = useState<string | null>(null);

  useEffect(() => {
    getEventConfigList({ pageSize: 1000, enabled: true }).then((res) => {
      if (res.code === 0 && res.data?.list) {
        setAllEventConfigs(res.data.list);
        setEventConfigs(
          res.data.list.map((item) => ({ label: item.name, value: item.id }))
        );
      }
    });
  }, []);

  const generateColumns = useCallback(
    (eventConfig: EventConfigWithFields): ProColumns<any>[] => {
      const columns: ProColumns<any>[] = [];

      // 根据字段配置生成列
      eventConfig.fields
        ?.filter((f) => f.isVisible) // 只显示可见字段
        .sort((a, b) => a.displayOrder - b.displayOrder) // 按显示顺序排序
        .forEach((field) => {
          columns.push({
            title: field.displayName,
            dataIndex: field.fieldName,
            key: field.fieldName,
            ellipsis: true,
          });
        });

      return columns;
    },
    []
  );

  const handleSearch = async (values: QueryEventDataRequest) => {
    setTableLoading(true);
    setCurrentEventId(values.eventId);
    try {
      const selectedConfig = allEventConfigs.find(
        (c) => c.id === values.eventId
      );
      if (selectedConfig) {
        setTableColumns(generateColumns(selectedConfig));
      }

      const response = await queryEventData(values.eventId, {
        ...values,
        page: pagination.current,
        pageSize: pagination.pageSize,
      });

      if (response.code === 0 && response.data) {
        setTableData(response.data.records || []);
        setTotal(response.data.total || 0);
      } else {
        message.error(response.msg || "查询失败");
      }
    } catch (_error) {
      message.error("查询失败");
    } finally {
      setTableLoading(false);
    }
  };

  const handleViewDetail = async (journeyId: number) => {
    if (!currentEventId) return;
    setDrawerVisible(true);
    setDetailLoading(true);
    try {
      const response = await getEventDetail(currentEventId, journeyId);
      if (response.code === 0) {
        setEventDetail(response.data);
      } else {
        message.error(response.msg || "获取详情失败");
      }
    } catch (_error) {
      message.error("获取详情失败");
    }
    setDetailLoading(false);
  };

  return {
    eventConfigs,
    tableColumns,
    tableData,
    tableLoading,
    total,
    pagination,
    drawerVisible,
    eventDetail,
    detailLoading,
    setPagination,
    setDrawerVisible,
    handleSearch,
    handleViewDetail,
  };
};
