/**
 * 逻辑图详情数据管理Hook
 */

import type { GraphConfigDetail, NodeConfig, EdgeConfig } from '@/services/lynxmanager/types';
import { getGraphConfig, updateGraphConfig } from '@/services/lynxmanager/api';
import { App } from 'antd';
import { useState, useEffect } from 'react';

export const useGraphDetail = (graphId: string) => {
  const { message } = App.useApp();
  const [data, setData] = useState<GraphConfigDetail | null>(null);
  const [loading, setLoading] = useState(false);

  // 获取详情
  const fetchDetail = async () => {
    if (!graphId) return;

    setLoading(true);
    try {
      const res = await getGraphConfig(graphId);
      if (res.code === 0 && res.data) {
        setData(res.data);
      } else {
        message.error(res.message || '获取详情失败');
      }
    } catch (error) {
      message.error('网络错误');
      console.error('获取逻辑图详情失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 更新基本信息
  const updateBasicInfo = async (values: Partial<GraphConfigDetail> & { tagIds?: string[] }) => {
    if (!data) return false;

    try {
      const res = await updateGraphConfig(
        graphId,
        {
          tenantId: data.tenantId,
          name: values.name || data.name,
          version: values.version || data.version,
          description: values.description || data.description,
          // Phase 2.5: 优先使用 tagIds，否则从 tags 对象数组转换
          tagIds: values.tagIds !== undefined ? values.tagIds : (values.tags || data.tags)?.map(tag => tag.id),
          isEnabled: values.isEnabled ?? data.isEnabled,
          icon: values.icon !== undefined ? values.icon : data.icon,
          iconColor: values.iconColor !== undefined ? values.iconColor : data.iconColor,
          nodes: data.nodes,
          edges: data.edges,
        }
      );

      if (res.code === 0) {
        message.success('更新成功');
        await fetchDetail(); // 刷新数据
        return true;
      }
      message.error(res.message || '更新失败');
      return false;
    } catch (error) {
      message.error('网络错误');
      console.error('更新基本信息失败:', error);
      return false;
    }
  };

  // 更新图配置（节点和边）
  const updateGraphStructure = async (nodes: NodeConfig[], edges: EdgeConfig[]) => {
    if (!data) return false;

    try {
      const res = await updateGraphConfig(
        graphId,
        {
          tenantId: data.tenantId,
          name: data.name,
          version: data.version,
          description: data.description,
          // Phase 2.5: 将 LynxTagSummary[] 转换为 string[] ID 数组
          tagIds: data.tags?.map(tag => tag.id),
          isEnabled: data.isEnabled,
          icon: data.icon,
          iconColor: data.iconColor,
          nodes,
          edges,
        }
      );

      if (res.code === 0) {
        message.success('保存成功');
        await fetchDetail();
        return true;
      }
      message.error(res.message || '保存失败');
      return false;
    } catch (error) {
      message.error('网络错误');
      console.error('更新图配置失败:', error);
      return false;
    }
  };

  // 初始化加载
  useEffect(() => {
    fetchDetail();
  }, [graphId]);

  return {
    data,
    loading,
    refresh: fetchDetail,
    updateBasicInfo,
    updateGraphStructure,
  };
};
