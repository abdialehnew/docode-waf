import React, { createContext, useContext, useState } from 'react';

const TabsContext = createContext();

export const Tabs = ({ children, value, onValueChange, className = '' }) => {
  const [activeTab, setActiveTab] = useState(value);

  const handleTabChange = (newValue) => {
    setActiveTab(newValue);
    if (onValueChange) {
      onValueChange(newValue);
    }
  };

  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab: handleTabChange }}>
      <div className={className}>
        {children}
      </div>
    </TabsContext.Provider>
  );
};

export const TabsList = ({ children, className = '' }) => {
  return (
    <div className={`flex border-b border-gray-200 ${className}`}>
      {children}
    </div>
  );
};

export const TabsTrigger = ({ children, value, className = '' }) => {
  const { activeTab, setActiveTab } = useContext(TabsContext);
  const isActive = activeTab === value;

  return (
    <button
      onClick={() => setActiveTab(value)}
      className={`px-4 py-2 font-medium transition-colors ${
        isActive
          ? 'text-blue-600 border-b-2 border-blue-600'
          : 'text-gray-600 hover:text-gray-900'
      } ${className}`}
    >
      {children}
    </button>
  );
};

export const TabsContent = ({ children, value, className = '' }) => {
  const { activeTab } = useContext(TabsContext);
  
  if (activeTab !== value) {
    return null;
  }

  return (
    <div className={`pt-4 ${className}`}>
      {children}
    </div>
  );
};
