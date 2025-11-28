/**
 * 图标选择器组件
 */

import { theme } from 'antd';
import { Empty, Input, Tabs } from 'antd';
import React, { useMemo, useState } from 'react';
import { getIconsByCategory, getColorByKey, type IconOption } from '../iconConfig';
import styles from './IconSelector.less';

interface IconSelectorProps {
  value?: string; // 当前选中的图标名称
  onChange?: (iconName: string) => void;
}

/**
 * 图标选择器
 */
const IconSelector: React.FC<IconSelectorProps> = ({ value, onChange }) => {
  const { token } = theme.useToken();
  const [searchText, setSearchText] = useState('');
  const iconsByCategory = useMemo(() => getIconsByCategory(), []);

  // 过滤图标
  const filterIcons = (icons: IconOption[]) => {
    if (!searchText) return icons;
    return icons.filter((icon) =>
      icon.label.toLowerCase().includes(searchText.toLowerCase()) ||
      icon.name.toLowerCase().includes(searchText.toLowerCase())
    );
  };

  // 渲染图标项
  const renderIconItem = (icon: IconOption) => {
    const IconComponent = icon.component;
    const isSelected = value === icon.name;

    return (
      <div
        key={icon.name}
        className={`${styles.iconItem} ${isSelected ? styles.selected : ''}`}
        onClick={() => onChange?.(icon.name)}
      >
        <div
          className={styles.iconBox}
          style={{
            borderColor: isSelected ? token.colorPrimary : 'transparent',
            backgroundColor: isSelected ? token.colorPrimaryBg : 'transparent',
          }}
        >
          <IconComponent size={32} color={icon.colorKey ? getColorByKey(icon.colorKey, token) : token.colorPrimary} />
        </div>
        <div className={styles.iconLabel}>{icon.label}</div>
      </div>
    );
  };

  // 创建 Tab 项
  const tabItems = Object.entries(iconsByCategory).map(([category, icons]) => ({
    key: category,
    label: category,
    children: (
      <div className={styles.iconGrid}>
        {filterIcons(icons).length > 0 ? (
          filterIcons(icons).map(renderIconItem)
        ) : (
          <Empty description="未找到匹配的图标" />
        )}
      </div>
    ),
  }));

  return (
    <div className={styles.iconSelector}>
      <div className={styles.searchBar}>
        <Input.Search
          placeholder="搜索图标..."
          allowClear
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
        />
      </div>
      <Tabs items={tabItems} />
    </div>
  );
};

export default IconSelector;
