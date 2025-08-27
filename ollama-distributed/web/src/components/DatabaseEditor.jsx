import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Modal, Form, Alert } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faDatabase,
  faPlus,
  faEdit,
  faTrash,
  faSearch,
  faSave,
  faUndo,
  faFileExport,
  faFileImport,
  faRefresh
} from '@fortawesome/free-solid-svg-icons';

const DatabaseEditor = ({ 
  tables = [], 
  onQuery, 
  onInsert, 
  onUpdate, 
  onDelete, 
  onExport, 
  onImport 
}) => {
  const [selectedTable, setSelectedTable] = useState(null);
  const [tableData, setTableData] = useState([]);
  const [showModal, setShowModal] = useState(false);
  const [modalType, setModalType] = useState('add'); // add, edit, query
  const [currentRow, setCurrentRow] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(25);

  useEffect(() => {
    if (selectedTable) {
      loadTableData();
    }
  }, [selectedTable, currentPage]);

  const loadTableData = async () => {
    setLoading(true);
    setError('');
    try {
      const offset = (currentPage - 1) * itemsPerPage;
      const result = await onQuery(`SELECT * FROM ${selectedTable} LIMIT ${itemsPerPage} OFFSET ${offset}`);
      setTableData(result.rows || []);
    } catch (err) {
      setError('Failed to load table data: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleRowEdit = (row) => {
    setCurrentRow(row);
    setModalType('edit');
    setShowModal(true);
  };

  const handleRowAdd = () => {
    setCurrentRow({});
    setModalType('add');
    setShowModal(true);
  };

  const handleRowDelete = async (row) => {
    if (window.confirm('Are you sure you want to delete this record?')) {
      try {
        await onDelete(selectedTable, row);
        await loadTableData();
      } catch (err) {
        setError('Failed to delete record: ' + err.message);
      }
    }
  };

  const handleSave = async (formData) => {
    try {
      if (modalType === 'add') {
        await onInsert(selectedTable, formData);
      } else {
        await onUpdate(selectedTable, currentRow.id, formData);
      }
      await loadTableData();
      setShowModal(false);
    } catch (err) {
      setError(`Failed to ${modalType} record: ` + err.message);
    }
  };

  const filteredData = tableData.filter(row =>
    Object.values(row).some(value =>
      value?.toString().toLowerCase().includes(searchTerm.toLowerCase())
    )
  );

  const getColumnNames = () => {
    if (tableData.length === 0) return [];
    return Object.keys(tableData[0]);
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Database Editor</h2>
        <div className="d-flex gap-2">
          <Button variant="outline-primary" onClick={() => onExport && onExport(selectedTable)}>
            <FontAwesomeIcon icon={faFileExport} className="me-2" />
            Export
          </Button>
          <Button variant="outline-success" onClick={() => onImport && onImport()}>
            <FontAwesomeIcon icon={faFileImport} className="me-2" />
            Import
          </Button>
          <Button variant="outline-info" onClick={loadTableData} disabled={!selectedTable}>
            <FontAwesomeIcon icon={faRefresh} className="me-2" />
            Refresh
          </Button>
        </div>
      </div>

      {error && (
        <Alert variant="danger" dismissible onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      <div className="row">
        {/* Table List */}
        <div className="col-md-3">
          <Card>
            <Card.Header>
              <h6 className="mb-0">
                <FontAwesomeIcon icon={faDatabase} className="me-2" />
                Tables
              </h6>
            </Card.Header>
            <Card.Body className="p-0">
              <div className="table-list">
                {tables.map(table => (
                  <div
                    key={table}
                    className={`table-item ${selectedTable === table ? 'active' : ''}`}
                    onClick={() => setSelectedTable(table)}
                  >
                    {table}
                  </div>
                ))}
              </div>
            </Card.Body>
          </Card>
        </div>

        {/* Table Data */}
        <div className="col-md-9">
          {selectedTable ? (
            <Card>
              <Card.Header>
                <div className="d-flex justify-content-between align-items-center">
                  <h6 className="mb-0">Table: {selectedTable}</h6>
                  <div className="d-flex gap-2">
                    <Form.Control
                      type="text"
                      placeholder="Search..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="search-input"
                    />
                    <Button variant="primary" size="sm" onClick={handleRowAdd}>
                      <FontAwesomeIcon icon={faPlus} className="me-1" />
                      Add Row
                    </Button>
                  </div>
                </div>
              </Card.Header>
              <Card.Body className="p-0">
                {loading ? (
                  <div className="text-center p-4">
                    <FontAwesomeIcon icon={faRefresh} spin className="me-2" />
                    Loading...
                  </div>
                ) : (
                  <div className="table-responsive">
                    <Table striped hover className="mb-0">
                      <thead>
                        <tr>
                          {getColumnNames().map(column => (
                            <th key={column}>{column}</th>
                          ))}
                          <th width="120">Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredData.map((row, index) => (
                          <tr key={index}>
                            {getColumnNames().map(column => (
                              <td key={column}>
                                {row[column]?.toString().substring(0, 50)}
                                {row[column]?.toString().length > 50 && '...'}
                              </td>
                            ))}
                            <td>
                              <div className="d-flex gap-1">
                                <Button
                                  variant="outline-primary"
                                  size="sm"
                                  onClick={() => handleRowEdit(row)}
                                >
                                  <FontAwesomeIcon icon={faEdit} />
                                </Button>
                                <Button
                                  variant="outline-danger"
                                  size="sm"
                                  onClick={() => handleRowDelete(row)}
                                >
                                  <FontAwesomeIcon icon={faTrash} />
                                </Button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  </div>
                )}
              </Card.Body>
            </Card>
          ) : (
            <Card>
              <Card.Body className="text-center">
                <FontAwesomeIcon icon={faDatabase} size="3x" className="text-muted mb-3" />
                <h5>Select a table to view and edit data</h5>
                <p className="text-muted">Choose a table from the left panel to start editing.</p>
              </Card.Body>
            </Card>
          )}
        </div>
      </div>

      {/* Add/Edit Modal */}
      <Modal show={showModal} onHide={() => setShowModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            <FontAwesomeIcon icon={modalType === 'add' ? faPlus : faEdit} className="me-2" />
            {modalType === 'add' ? 'Add New Record' : 'Edit Record'}
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={(e) => {
          e.preventDefault();
          const formData = new FormData(e.target);
          const data = Object.fromEntries(formData.entries());
          handleSave(data);
        }}>
          <Modal.Body>
            <div className="row">
              {getColumnNames().map(column => (
                <div key={column} className="col-md-6 mb-3">
                  <Form.Label>{column}</Form.Label>
                  <Form.Control
                    name={column}
                    defaultValue={currentRow[column] || ''}
                    disabled={column === 'id' && modalType === 'edit'}
                  />
                </div>
              ))}
            </div>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              <FontAwesomeIcon icon={faSave} className="me-2" />
              Save
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default DatabaseEditor;